package repository

import (
	"context"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
)

type RatingRepository struct{}

func NewRatingRepository() *RatingRepository {
	return &RatingRepository{}
}

func (r *RatingRepository) SaveRating(ctx context.Context, rating domain.Rating) error {
	//data, err := r.fw.Read()
	//if err != nil {
	//	return fmt.Errorf("r.fw.Read: %w", err)
	//}
	//
	//var ratings []domain.Rating
	//err = json.Unmarshal(data, &ratings)
	//if err != nil {
	//	ratings = make([]domain.Rating, 0)
	//}
	//
	//ratings = append(ratings, rating)
	//
	//b, err := json.Marshal(ratings)
	//if err != nil {
	//	return fmt.Errorf("json.Marshal: %w", err)
	//}
	//
	//_, err = r.fw.TruncateAndWrite(b)
	//if err != nil {
	//	return fmt.Errorf("r.fw.TruncateAndWrite: %w", err)
	//}

	return nil
}
