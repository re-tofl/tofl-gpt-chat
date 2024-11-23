package usecase

import (
	"context"
)

type SpeechStore interface {
	SpeechToText(filePath string) string
}

type SpeechUsecase struct {
	store SpeechStore
}

func NewSpeechUsecase(store SpeechStore) *SpeechUsecase {
	return &SpeechUsecase{
		store: store,
	}
}

func (s *SpeechUsecase) ConvertSpeechToText(ctx context.Context, filePath string) string {
	answer := s.store.SpeechToText(filePath)
	return answer
}
