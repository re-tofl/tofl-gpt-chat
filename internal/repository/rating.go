package repository

import "io"

type RatingRepository struct {
	fileWriter io.WriteCloser
}

func NewRatingRepository(fileWriter io.WriteCloser) *RatingRepository {
	return &RatingRepository{
		fileWriter: fileWriter,
	}
}
