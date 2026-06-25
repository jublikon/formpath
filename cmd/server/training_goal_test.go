package main

import (
	"math"
	"strings"
	"testing"
)

func validTrainingGoal() TrainingGoal {
	targetDurationSeconds := 3*60*60 + 45*60
	return TrainingGoal{
		UserID:                "00000000-0000-0000-0000-000000000001",
		GoalType:              trainingGoalTypeDistanceEvent,
		Sport:                 trainingGoalSportRun,
		Name:                  "Berlin Marathon",
		TargetDistanceMeters:  42195,
		TargetDate:            "2026-09-27",
		TargetDurationSeconds: &targetDurationSeconds,
	}
}

func TestTrainingGoalValidate_AcceptsValidGoals(t *testing.T) {
	tests := []struct {
		name string
		goal TrainingGoal
	}{
		{
			name: "running goal with target duration",
			goal: validTrainingGoal(),
		},
		{
			name: "cycling goal without target duration",
			goal: func() TrainingGoal {
				goal := validTrainingGoal()
				goal.Sport = trainingGoalSportRide
				goal.Name = "Summer Century"
				goal.TargetDistanceMeters = 160934
				goal.TargetDurationSeconds = nil
				return goal
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.goal.Validate(); err != nil {
				t.Fatalf("Expected valid goal, got %v", err)
			}
		})
	}
}

func TestTrainingGoalValidate_RejectsInvalidGoals(t *testing.T) {
	tests := []struct {
		name    string
		change  func(*TrainingGoal)
		wantErr string
	}{
		{
			name: "unsupported goal type",
			change: func(goal *TrainingGoal) {
				goal.GoalType = "capability"
			},
			wantErr: "goal type must be distance_event",
		},
		{
			name: "unsupported sport",
			change: func(goal *TrainingGoal) {
				goal.Sport = "swim"
			},
			wantErr: "sport must be run or ride",
		},
		{
			name: "blank name",
			change: func(goal *TrainingGoal) {
				goal.Name = "   "
			},
			wantErr: "name is required",
		},
		{
			name: "zero target distance",
			change: func(goal *TrainingGoal) {
				goal.TargetDistanceMeters = 0
			},
			wantErr: "target distance must be greater than zero",
		},
		{
			name: "negative target distance",
			change: func(goal *TrainingGoal) {
				goal.TargetDistanceMeters = -1
			},
			wantErr: "target distance must be greater than zero",
		},
		{
			name: "NaN target distance",
			change: func(goal *TrainingGoal) {
				goal.TargetDistanceMeters = math.NaN()
			},
			wantErr: "target distance must be finite",
		},
		{
			name: "infinite target distance",
			change: func(goal *TrainingGoal) {
				goal.TargetDistanceMeters = math.Inf(1)
			},
			wantErr: "target distance must be finite",
		},
		{
			name: "malformed target date",
			change: func(goal *TrainingGoal) {
				goal.TargetDate = "27-09-2026"
			},
			wantErr: "target date must be a valid YYYY-MM-DD calendar date",
		},
		{
			name: "impossible target date",
			change: func(goal *TrainingGoal) {
				goal.TargetDate = "2026-02-30"
			},
			wantErr: "target date must be a valid YYYY-MM-DD calendar date",
		},
		{
			name: "zero target duration",
			change: func(goal *TrainingGoal) {
				value := 0
				goal.TargetDurationSeconds = &value
			},
			wantErr: "target duration must be greater than zero",
		},
		{
			name: "negative target duration",
			change: func(goal *TrainingGoal) {
				value := -1
				goal.TargetDurationSeconds = &value
			},
			wantErr: "target duration must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goal := validTrainingGoal()
			tt.change(&goal)

			err := goal.Validate()
			if err == nil {
				t.Fatal("Expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}
