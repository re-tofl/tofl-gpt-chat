package usecase

import (
	"context"

	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
)

type SearchStore interface {
	LoadJSONArrayFromFile(path string) []domain.DatabaseItem
	DoDatabaseEmbedding(ctx context.Context, items []domain.DatabaseItem)
}

type SearchUseCase struct {
	store SearchStore
}

func NewSearchUseCase(store SearchStore) *SearchUseCase {
	return &SearchUseCase{
		store: store,
	}
}

func (u *SearchUseCase) DatabaseToVector(ctx context.Context) {
	items := u.store.LoadJSONArrayFromFile("path")
	u.store.DoDatabaseEmbedding(ctx, items)
}
