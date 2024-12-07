package repository

import (
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"io"
)

type RatingRepository struct {
	fileWriter io.WriteCloser
}

func NewRatingRepository(fileWriter io.WriteCloser) *RatingRepository {
	return &RatingRepository{
		fileWriter: fileWriter,
	}
}

func (r *RatingRepository) Save(rating domain.Rating) error {
	return nil
}
