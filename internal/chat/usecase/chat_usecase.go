package usecase

import "tgbot/domain"

type ChatStore interface {
	FullTextSearch(originalMessageText string) []domain.Chunk
}

func SendMessageFromUserToLLM(chatStore ChatStore, userMessage domain.Message) {
	userMessage.MessageContext = chatStore.FullTextSearch(userMessage.OriginalMessageText)
	//...
}
