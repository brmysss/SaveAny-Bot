package bot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/krau/SaveAny-Bot/dao"
	"github.com/krau/SaveAny-Bot/logger"
	"github.com/krau/SaveAny-Bot/storage"
	"github.com/krau/SaveAny-Bot/types"
)

var (
	linkRegexString = `t.me/.*/\d+`
	linkRegex       = regexp.MustCompile(linkRegexString)
)

func handleLinkMessage(ctx *ext.Context, update *ext.Update) error {
	logger.L.Trace("Got link message")
	link := linkRegex.FindString(update.EffectiveMessage.Text)
	if link == "" {
		return dispatcher.ContinueGroups
	}
	strSlice := strings.Split(link, "/")
	if len(strSlice) < 3 {
		return dispatcher.ContinueGroups
	}
	messageID, err := strconv.Atoi(strSlice[2])
	if err != nil {
		logger.L.Errorf("解析消息 ID 失败: %s", err)
		ctx.Reply(update, ext.ReplyTextString("无法解析消息 ID"), nil)
		return dispatcher.EndGroups
	}
	chatUsername := strSlice[1]
	linkChat, err := ctx.ResolveUsername(chatUsername)
	if err != nil {
		logger.L.Errorf("解析 Chat ID 失败: %s", err)
		ctx.Reply(update, ext.ReplyTextString("无法解析 Chat ID"), nil)
		return dispatcher.EndGroups
	}
	if linkChat == nil {
		logger.L.Errorf("无法找到聊天: %s", chatUsername)
		ctx.Reply(update, ext.ReplyTextString("无法找到聊天"), nil)
		return dispatcher.EndGroups
	}
	user, err := dao.GetUserByChatID(update.GetUserChat().GetID())
	if err != nil {
		logger.L.Errorf("获取用户失败: %s", err)
		ctx.Reply(update, ext.ReplyTextString("获取用户失败"), nil)
		return dispatcher.EndGroups
	}
	storages := storage.GetUserStorages(user.ChatID)

	if len(storages) == 0 {
		ctx.Reply(update, ext.ReplyTextString("无可用的存储"), nil)
		return dispatcher.EndGroups
	}
	replied, err := ctx.Reply(update, ext.ReplyTextString("正在获取文件..."), nil)
	if err != nil {
		logger.L.Errorf("回复失败: %s", err)
		return dispatcher.EndGroups
	}

	file, err := FileFromMessage(ctx, linkChat.GetID(), messageID, "")
	if err != nil {
		logger.L.Errorf("获取文件失败: %s", err)
		ctx.Reply(update, ext.ReplyTextString("获取文件失败: "+err.Error()), nil)
		return dispatcher.EndGroups
	}
	// TODO: Better file name
	if file.FileName == "" {
		logger.L.Warnf("文件名为空，使用生成的名称")
		file.FileName = fmt.Sprintf("%d_%d_%s", linkChat.GetID(), messageID, file.Hash())
	}

	receivedFile := &types.ReceivedFile{
		Processing:     false,
		FileName:       file.FileName,
		ChatID:         linkChat.GetID(),
		MessageID:      messageID,
		ReplyMessageID: replied.ID,
		ReplyChatID:    update.GetUserChat().GetID(),
	}
	if err := dao.SaveReceivedFile(receivedFile); err != nil {
		logger.L.Errorf("保存接收的文件失败: %s", err)
		ctx.EditMessage(update.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
			Message: "无法保存文件: " + err.Error(),
			ID:      replied.ID,
		})
		return dispatcher.EndGroups
	}
	if !user.Silent || user.DefaultStorage == "" {
		return ProvideSelectMessage(ctx, update, file, linkChat.GetID(), messageID, replied.ID)
	}
	return HandleSilentAddTask(ctx, update, user, &types.Task{
		Ctx:            ctx,
		Status:         types.Pending,
		File:           file,
		StorageName:    user.DefaultStorage,
		UserID:         user.ChatID,
		FileChatID:     linkChat.GetID(),
		FileMessageID:  messageID,
		ReplyMessageID: replied.ID,
		ReplyChatID:    update.GetUserChat().GetID(),
	})
}
