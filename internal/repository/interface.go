package repository

import (
	"context"
	"subs_service/internal/models"
	"time"

	"github.com/google/uuid"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *models.Subscription) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, sub *models.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]models.Subscription, error)
	GetTotalCost(ctx context.Context, userID uuid.UUID, serviceName string, from, to time.Time) (int, error)
}
