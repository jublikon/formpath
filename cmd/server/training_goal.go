package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	trainingGoalTypeDistanceEvent = "distance_event"
	trainingGoalSportRun          = "run"
	trainingGoalSportRide         = "ride"
	maxTrainingGoalDuration       = 2147483647
)

var ErrTrainingGoalNotFound = errors.New("training goal not found")

type TrainingGoal struct {
	ID                    string     `json:"id,omitempty"`
	UserID                string     `json:"user_id"`
	GoalType              string     `json:"goal_type"`
	Sport                 string     `json:"sport"`
	Name                  string     `json:"name"`
	TargetDistanceMeters  float64    `json:"target_distance_meters"`
	TargetDate            string     `json:"target_date"`
	TargetDurationSeconds *int       `json:"target_duration_seconds,omitempty"`
	CreatedAt             *time.Time `json:"created_at,omitempty"`
	UpdatedAt             *time.Time `json:"updated_at,omitempty"`
}

type TrainingGoalStore interface {
	SaveTrainingGoal(ctx context.Context, goal TrainingGoal) error
	GetTrainingGoal(ctx context.Context, userID string) (TrainingGoal, error)
	DeleteTrainingGoal(ctx context.Context, userID string) error
}

var trainingGoalStore TrainingGoalStore = noopTrainingGoalStore{}

type noopTrainingGoalStore struct{}

func (noopTrainingGoalStore) SaveTrainingGoal(ctx context.Context, goal TrainingGoal) error {
	return fmt.Errorf("training goal store is not configured")
}

func (noopTrainingGoalStore) GetTrainingGoal(ctx context.Context, userID string) (TrainingGoal, error) {
	return TrainingGoal{}, fmt.Errorf("training goal store is not configured")
}

func (noopTrainingGoalStore) DeleteTrainingGoal(ctx context.Context, userID string) error {
	return fmt.Errorf("training goal store is not configured")
}

type PostgresTrainingGoalStore struct {
	db *sql.DB
}

var _ TrainingGoalStore = (*PostgresTrainingGoalStore)(nil)

func NewPostgresTrainingGoalStore(db *sql.DB) *PostgresTrainingGoalStore {
	return &PostgresTrainingGoalStore{db: db}
}

func (store *PostgresTrainingGoalStore) SaveTrainingGoal(ctx context.Context, goal TrainingGoal) error {
	if err := goal.Validate(); err != nil {
		return fmt.Errorf("validating training goal: %w", err)
	}

	_, err := store.db.ExecContext(ctx, `
		insert into training_goals (
			user_id,
			goal_type,
			sport,
			name,
			target_distance_meters,
			target_date,
			target_duration_seconds,
			updated_at
		)
		values ($1, $2, $3, $4, $5, $6::date, $7, now())
		on conflict (user_id)
		do update set
			goal_type = excluded.goal_type,
			sport = excluded.sport,
			name = excluded.name,
			target_distance_meters = excluded.target_distance_meters,
			target_date = excluded.target_date,
			target_duration_seconds = excluded.target_duration_seconds,
			updated_at = now()
	`, goal.UserID, goal.GoalType, goal.Sport, strings.TrimSpace(goal.Name), goal.TargetDistanceMeters, goal.TargetDate, nullableInt(goal.TargetDurationSeconds))
	if err != nil {
		return fmt.Errorf("saving training goal: %w", err)
	}
	return nil
}

func (store *PostgresTrainingGoalStore) GetTrainingGoal(ctx context.Context, userID string) (TrainingGoal, error) {
	var goal TrainingGoal
	var targetDate time.Time
	var targetDuration sql.NullInt64
	var createdAt time.Time
	var updatedAt time.Time

	err := store.db.QueryRowContext(ctx, `
		select
			id::text,
			user_id::text,
			goal_type,
			sport,
			name,
			target_distance_meters,
			target_date,
			target_duration_seconds,
			created_at,
			updated_at
		from training_goals
		where user_id = $1
	`, userID).Scan(
		&goal.ID,
		&goal.UserID,
		&goal.GoalType,
		&goal.Sport,
		&goal.Name,
		&goal.TargetDistanceMeters,
		&targetDate,
		&targetDuration,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return TrainingGoal{}, ErrTrainingGoalNotFound
	}
	if err != nil {
		return TrainingGoal{}, fmt.Errorf("loading training goal: %w", err)
	}

	goal.TargetDate = targetDate.Format("2006-01-02")
	goal.TargetDurationSeconds = intPointerFromNull(targetDuration)
	goal.CreatedAt = &createdAt
	goal.UpdatedAt = &updatedAt
	return goal, nil
}

func (store *PostgresTrainingGoalStore) DeleteTrainingGoal(ctx context.Context, userID string) error {
	result, err := store.db.ExecContext(ctx, `
		delete from training_goals
		where user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("deleting training goal: %w", err)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking deleted training goal: %w", err)
	}
	if deleted == 0 {
		return ErrTrainingGoalNotFound
	}
	return nil
}

func (goal TrainingGoal) Validate() error {
	if goal.GoalType != trainingGoalTypeDistanceEvent {
		return errors.New("goal type must be distance_event")
	}
	if goal.Sport != trainingGoalSportRun && goal.Sport != trainingGoalSportRide {
		return errors.New("sport must be run or ride")
	}
	if strings.TrimSpace(goal.Name) == "" {
		return errors.New("name is required")
	}
	if math.IsNaN(goal.TargetDistanceMeters) || math.IsInf(goal.TargetDistanceMeters, 0) {
		return errors.New("target distance must be finite")
	}
	if goal.TargetDistanceMeters <= 0 {
		return errors.New("target distance must be greater than zero")
	}
	if !validCalendarDate(goal.TargetDate) {
		return errors.New("target date must be a valid YYYY-MM-DD calendar date")
	}
	if goal.TargetDurationSeconds != nil {
		if *goal.TargetDurationSeconds <= 0 {
			return errors.New("target duration must be greater than zero")
		}
		if *goal.TargetDurationSeconds > maxTrainingGoalDuration {
			return errors.New("target duration is too large")
		}
	}
	return nil
}

func validCalendarDate(value string) bool {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return false
	}
	return parsed.Format("2006-01-02") == value
}

func nullableInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func intPointerFromNull(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	converted := int(value.Int64)
	return &converted
}
