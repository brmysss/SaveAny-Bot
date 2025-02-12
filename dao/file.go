package dao

import "github.com/krau/SaveAny-Bot/types"

func SaveReceivedFile(receivedFile *types.ReceivedFile) error {
	record, err := GetReceivedFileByChatAndMessageID(receivedFile.ChatID, receivedFile.MessageID)
	if err == nil {
		receivedFile.ID = record.ID
	}
	return db.Save(receivedFile).Error
}

func GetReceivedFileByChatAndMessageID(chatID int64, messageID int) (*types.ReceivedFile, error) {
	var receivedFile types.ReceivedFile
	err := db.Where("chat_id = ? AND message_id = ?", chatID, messageID).First(&receivedFile).Error
	if err != nil {
		return nil, err
	}
	return &receivedFile, nil
}

func DeleteReceivedFile(receivedFile *types.ReceivedFile) error {
	return db.Delete(receivedFile).Error
}
