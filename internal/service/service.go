package service

import (
	"context"
	"errors"
	"fmt"
	"subs_service/internal/models"
	"subs_service/internal/repository"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidDate  = errors.New("invalid subscription data")
	ErrInvalidName  = errors.New("invalid subscription name")
	ErrInvalidPrice = errors.New("invalid subscription price")
)

type subscriptionSrv struct {
	repo repository.SubscriptionRepository
}

func NewSubscriptionService(repo repository.SubscriptionRepository) *subscriptionSrv {
	return &subscriptionSrv{
		repo: repo,
	}
}

func (s *subscriptionSrv) Create(ctx context.Context, sub *models.Subscription) (uuid.UUID, error) {
	if sub.ServiceName == "" {
		return uuid.Nil, fmt.Errorf("service.Create name is empty %w", ErrInvalidName)
	}
	if sub.Price < 0 {
		return uuid.Nil, fmt.Errorf("service.Create price is negative %w", ErrInvalidPrice)
	}
	if sub.EndDate != nil {
		if sub.EndDate.Before(sub.StartDate) {
			return uuid.Nil, fmt.Errorf("service.Create end date is in the future %w", ErrInvalidDate)
		}
	}

	return s.repo.Create(ctx, sub)
}
func (s *subscriptionSrv) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *subscriptionSrv) Update(ctx context.Context, sub *models.Subscription) error {
	if sub.ServiceName == "" {
		return fmt.Errorf("service.Update name is empty %w", ErrInvalidName)
	}
	if sub.Price < 0 {
		return fmt.Errorf("service.Update price is negative %w", ErrInvalidPrice)
	}
	if sub.EndDate != nil {
		if sub.EndDate.Before(sub.StartDate) {
			return fmt.Errorf("service.Create end date is in the future %w", ErrInvalidDate)
		}
	}
	return s.repo.Update(ctx, sub)
}
func (s *subscriptionSrv) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
func (s *subscriptionSrv) List(ctx context.Context) ([]models.Subscription, error) {
	return s.repo.List(ctx)
}
func (s *subscriptionSrv) CalculateTotalCost(ctx context.Context, userID uuid.UUID, serviceName string, from, to time.Time) (int, error) {
	subs, err := s.repo.GetTotalCost(ctx, userID, serviceName, from, to)
	if err != nil {
		return 0, fmt.Errorf("service.CalculateTotalCost: %w", err)
	}

	var totalCost int

	for _, sub := range subs {
		overlapStart := sub.StartDate
		if from.After(overlapStart) {
			overlapStart = from
		}

		overlapEnd := to
		if sub.EndDate != nil && sub.EndDate.Before(to) {
			overlapEnd = *sub.EndDate
		}

		if !overlapStart.After(overlapEnd) {
			years := overlapEnd.Year() - overlapStart.Year()
			months := int(overlapEnd.Month()) - int(overlapStart.Month())

			totalMonths := years*12 + months + 1

			if totalMonths > 0 {
				totalCost += totalMonths * sub.Price
			}
		}
	}

	return totalCost, nil
}
