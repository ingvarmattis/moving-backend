package reviews

import (
	"context"
	"errors"
	"fmt"

	"github.com/ingvarmattis/moving/src/services"
	reviewssvc "github.com/ingvarmattis/moving/src/services/reviews"
)

var ErrNotFound = errors.New("not found")

type Handlers struct {
	ReviewsService services.ReviewsService
}

func (s *Handlers) Reviews(ctx context.Context) ([]Review, error) {
	svcReviews, err := s.ReviewsService.Reviews(ctx)
	if err != nil {
		if errors.Is(err, reviewssvc.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed get reviews | %w", err)
	}

	reviews := make([]Review, 0, len(svcReviews))

	for _, svcReview := range svcReviews {
		reviews = append(reviews, Review{
			Text:     svcReview.Text,
			Name:     svcReview.Name,
			PhotoURL: svcReview.PhotoURL,
			Rate:     svcReview.Rate,
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
