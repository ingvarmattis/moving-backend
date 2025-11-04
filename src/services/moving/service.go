package moving

import (
	"context"
	"errors"
)

const serviceName = "moving-service"

var ErrNotFound = errors.New("not found")

//go:generate bash -c "mkdir -p mocks"
//go:generate mockgen -source=service.go -destination=mocks/mocks.go -package=mocks
type movingStorage interface {
}

type Service struct {
	movingStorage movingStorage
}

func NewService(ctx context.Context, movingStorage movingStorage) (*Service, error) {
	service := &Service{movingStorage: movingStorage}

	return service, nil
}
