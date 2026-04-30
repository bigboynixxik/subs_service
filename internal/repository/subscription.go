package repository

import (
	"context"
	"errors"
	"fmt"
	"subs_service/internal/models"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type SubscriptionRepo struct {
	db *pgxpool.Pool
	sq sq.StatementBuilderType
}

func NewSubscriptionRepo(db *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{
		db: db,
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *SubscriptionRepo) Create(ctx context.Context, sub *models.Subscription) (uuid.UUID, error) {
	query, args, err := r.sq.Insert("subscriptions").
		Columns("service_name", "price", "user_id", "start_date", "end_date").
		Values(sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("repository.Create subscription query: %w", err)
	}

	var id uuid.UUID
	err = r.db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("repository.Create subscription scan: %w", err)
	}
	return id, nil
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	query, args, err := r.sq.Select("id", "service_name", "price", "user_id", "start_date", "end_date").
		From("subscriptions").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("repository.GetByID query: %w", err)
	}
	var sub models.Subscription
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("repository.GetByID scan: %w", err)
	}
	return &sub, nil
}
func (r *SubscriptionRepo) Update(ctx context.Context, sub *models.Subscription) error {
	query, args, err := r.sq.Update("subscriptions").
		Set("service_name", sub.ServiceName).
		Set("price", sub.Price).
		Set("user_id", sub.UserID).
		Set("start_date", sub.StartDate).
		Set("end_date", sub.EndDate).
		Where(sq.Eq{"id": sub.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("repository.Update subscription query: %w", err)
	}
	cmd, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("repository.Update subscription scan: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}
	return nil
}

func (r *SubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := r.sq.Delete("subscriptions").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("repository.Delete subscription query: %w", err)
	}
	cmd, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("repository.Delete subscription scan: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}
	return nil
}

func (r *SubscriptionRepo) List(ctx context.Context) ([]models.Subscription, error) {
	query, args, err := r.sq.Select("id", "service_name", "price", "user_id", "start_date", "end_date").
		From("subscriptions").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("repository.List query: %w", err)
	}
	var subs []models.Subscription
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repository.List subscription scan: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var sub models.Subscription
		err = rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate)
		if err != nil {
			return nil, fmt.Errorf("repository.List scan: %w", err)
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.List scan: %w", err)
	}
	return subs, nil
}
func (r *SubscriptionRepo) GetTotalCost(ctx context.Context, userID uuid.UUID, serviceName string, from, to time.Time) (int, error) {
	queryBuilder := r.sq.Select("COALESCE(SUM(price), 0)").
		From("subscriptions").
		Where(sq.Eq{"user_id": userID}).
		Where(sq.LtOrEq{"start_date": to}).
		Where(sq.Or{
			sq.Eq{"end_date": nil},
			sq.GtOrEq{"end_date": from},
		})
	if serviceName != "" {
		queryBuilder = queryBuilder.Where(sq.Eq{"service_name": serviceName})
	}
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("repository.GetTotalCost query: %w", err)
	}
	var totalCost int
	err = r.db.QueryRow(ctx, query, args...).Scan(&totalCost)
	if err != nil {
		return 0, fmt.Errorf("repository.GetTotalCost scan: %w", err)
	}
	return totalCost, nil
}
