package reviews

import (
	"context"
	"errors"
	"fmt"

	repo "github.com/ingvarmattis/moving/src/repositories/reviews"
)

var ErrNotFound = errors.New("not found")

//go:generate bash -c "mkdir -p mocks"
//go:generate mockgen -source=service.go -destination=mocks/mocks.go -package=mocks
type reviewStorage interface {
	Reviews(ctx context.Context) ([]*repo.Review, error)
}

type Service struct {
	reviewStorage reviewStorage
}

func NewService(reviewStorage reviewStorage) *Service {
	return &Service{reviewStorage: reviewStorage}
}

func (s *Service) Reviews(ctx context.Context) ([]Review, error) {
	repoReviews, err := s.reviewStorage.Reviews(ctx)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get all reviews | %w", err)
	}

	reviews := make([]Review, 0, len(repoReviews))

	for _, repoOrder := range repoReviews {
		reviews = append(reviews, Review{
			Text:     repoOrder.Text,
			Name:     repoOrder.Name,
			PhotoURL: repoOrder.PhotoURL,
			Rate:     repoOrder.Rate,
		})
	}

	return reviews, nil
}

type Review struct {
	Text     string
	Name     string
	PhotoURL string
	Rate     int32
}
