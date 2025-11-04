package services

type SvcLayer struct {
	MovingService MovingService
}

type MovingService interface{}
