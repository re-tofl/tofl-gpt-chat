package repository

import (
	"context"
	"encoding/json"
	"github.com/re-tofl/tofl-gpt-chat/internal/adapters"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
)

type RatingRepository struct {
	fw *adapters.FileWriter
}

func NewRatingRepository(fw *adapters.FileWriter) *RatingRepository {
	return &RatingRepository{
		fw: fw,
	}
}

func (r *RatingRepository) SaveRating(ctx context.Context, rating domain.Rating) error {
	data, err := r.fw.Read()
	if err != nil {
		return err
	}

	var ratings []domain.Rating
	err = json.Unmarshal(data, &ratings)
	if err != nil {
		return err
	}

	ratings = append(ratings, domain.Rating{
		ContextID: rating.ContextID,
		Rating:    rating.Rating,
	})

	b, err := json.Marshal(ratings)
	if err != nil {
		return err
	}

	_, err = r.fw.Write(b)
	if err != nil {
		return err
	}

	return nil
}
