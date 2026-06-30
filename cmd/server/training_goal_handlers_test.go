package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeTrainingGoalStore struct {
	goal        TrainingGoal
	exists      bool
	saveErr     error
	getErr      error
	deleteErr   error
	saveCount   int
	getCount    int
	deleteCount int
}

func (store *fakeTrainingGoalStore) SaveTrainingGoal(ctx context.Context, goal TrainingGoal) error {
	store.saveCount++
	if store.saveErr != nil {
		return store.saveErr
	}
	store.goal = goal
	store.exists = true
	return nil
}

func (store *fakeTrainingGoalStore) GetTrainingGoal(ctx context.Context, userID string) (TrainingGoal, error) {
	store.getCount++
	if store.getErr != nil {
		return TrainingGoal{}, store.getErr
	}
	if !store.exists {
		return TrainingGoal{}, ErrTrainingGoalNotFound
	}
	return store.goal, nil
}

func (store *fakeTrainingGoalStore) DeleteTrainingGoal(ctx context.Context, userID string) error {
	store.deleteCount++
	if store.deleteErr != nil {
		return store.deleteErr
	}
	if !store.exists {
		return ErrTrainingGoalNotFound
	}
	store.exists = false
	return nil
}

func withTrainingGoalStore(t *testing.T, store TrainingGoalStore) {
	t.Helper()

	originalTrainingGoalStore := trainingGoalStore
	t.Cleanup(func() {
		trainingGoalStore = originalTrainingGoalStore
	})
	trainingGoalStore = store
}

func TestTrainingGoalGetHandler_ReturnsStoredGoal(t *testing.T) {
	t.Setenv("APP_USER_ID", "00000000-0000-0000-0000-000000000042")

	targetDurationSeconds := 3*60*60 + 30*60
	store := &fakeTrainingGoalStore{
		exists: true,
		goal: TrainingGoal{
			ID:                    "goal-1",
			UserID:                "00000000-0000-0000-0000-000000000042",
			GoalType:              trainingGoalTypeDistanceEvent,
			Sport:                 trainingGoalSportRun,
			Name:                  "Berlin Marathon",
			TargetDistanceMeters:  42195,
			TargetDate:            "2026-09-27",
			TargetDurationSeconds: &targetDurationSeconds,
		},
	}
	withTrainingGoalStore(t, store)

	req := httptest.NewRequest(http.MethodGet, "/api/training-goal", nil)
	w := httptest.NewRecorder()

	trainingGoalGetHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200 OK, got %d: %s", resp.StatusCode, string(body))
	}

	var got TrainingGoal
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if got.Name != "Berlin Marathon" {
		t.Fatalf("Expected Berlin Marathon goal, got %q", got.Name)
	}
	if got.UserID != "00000000-0000-0000-0000-000000000042" {
		t.Fatalf("Expected configured app user ID, got %q", got.UserID)
	}
	if got.TargetDurationSeconds == nil || *got.TargetDurationSeconds != targetDurationSeconds {
		t.Fatalf("Expected target duration %d, got %v", targetDurationSeconds, got.TargetDurationSeconds)
	}
}

func TestTrainingGoalGetHandler_ReturnsNotFoundWhenNoGoalExists(t *testing.T) {
	withTrainingGoalStore(t, &fakeTrainingGoalStore{})

	req := httptest.NewRequest(http.MethodGet, "/api/training-goal", nil)
	w := httptest.NewRecorder()

	trainingGoalGetHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404 Not Found, got %d", resp.StatusCode)
	}
}

func TestTrainingGoalGetHandler_ReturnsInternalServerErrorWhenStoreFails(t *testing.T) {
	withTrainingGoalStore(t, &fakeTrainingGoalStore{getErr: errors.New("database unavailable")})

	req := httptest.NewRequest(http.MethodGet, "/api/training-goal", nil)
	w := httptest.NewRecorder()

	trainingGoalGetHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected status 500 Internal Server Error, got %d", resp.StatusCode)
	}
}

func TestTrainingGoalPutHandler_SavesGoalWithServerOwnedFields(t *testing.T) {
	t.Setenv("APP_USER_ID", "00000000-0000-0000-0000-000000000042")

	store := &fakeTrainingGoalStore{}
	withTrainingGoalStore(t, store)

	body := []byte(`{
		"user_id": "00000000-0000-0000-0000-000000000999",
		"goal_type": "capability",
		"sport": "run",
		"name": "Berlin Marathon",
		"target_distance_meters": 42195,
		"target_date": "2026-09-27",
		"target_duration_seconds": 12600
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/training-goal", bytes.NewReader(body))
	w := httptest.NewRecorder()

	trainingGoalPutHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200 OK, got %d: %s", resp.StatusCode, string(body))
	}
	if store.saveCount != 1 {
		t.Fatalf("Expected one save, got %d", store.saveCount)
	}
	if store.getCount != 1 {
		t.Fatalf("Expected saved goal to be loaded once, got %d loads", store.getCount)
	}
	if store.goal.UserID != "00000000-0000-0000-0000-000000000042" {
		t.Fatalf("Expected handler to set configured app user ID, got %q", store.goal.UserID)
	}
	if store.goal.GoalType != trainingGoalTypeDistanceEvent {
		t.Fatalf("Expected handler to set distance event goal type, got %q", store.goal.GoalType)
	}

	var got TrainingGoal
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if got.UserID != store.goal.UserID {
		t.Fatalf("Expected response user ID %q, got %q", store.goal.UserID, got.UserID)
	}
}

func TestTrainingGoalPutHandler_ReturnsBadRequestForInvalidGoal(t *testing.T) {
	store := &fakeTrainingGoalStore{}
	withTrainingGoalStore(t, store)

	body := []byte(`{
		"sport": "run",
		"name": "   ",
		"target_distance_meters": 42195,
		"target_date": "2026-09-27"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/training-goal", bytes.NewReader(body))
	w := httptest.NewRecorder()

	trainingGoalPutHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status 400 Bad Request, got %d", resp.StatusCode)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(bodyBytes), "name is required") {
		t.Fatalf("Expected validation message in response, got %q", string(bodyBytes))
	}
	if store.saveCount != 0 {
		t.Fatalf("Expected invalid goal not to be saved, got %d saves", store.saveCount)
	}
}

func TestTrainingGoalPutHandler_ReturnsBadRequestForInvalidJSON(t *testing.T) {
	store := &fakeTrainingGoalStore{}
	withTrainingGoalStore(t, store)

	req := httptest.NewRequest(http.MethodPut, "/api/training-goal", strings.NewReader(`{`))
	w := httptest.NewRecorder()

	trainingGoalPutHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status 400 Bad Request, got %d", resp.StatusCode)
	}
	if store.saveCount != 0 {
		t.Fatalf("Expected invalid JSON not to be saved, got %d saves", store.saveCount)
	}
}

func TestTrainingGoalPutHandler_ReturnsInternalServerErrorWhenStoreFails(t *testing.T) {
	store := &fakeTrainingGoalStore{saveErr: errors.New("database unavailable")}
	withTrainingGoalStore(t, store)

	body := []byte(`{
		"sport": "ride",
		"name": "Summer Century",
		"target_distance_meters": 160934,
		"target_date": "2026-07-12"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/training-goal", bytes.NewReader(body))
	w := httptest.NewRecorder()

	trainingGoalPutHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected status 500 Internal Server Error, got %d", resp.StatusCode)
	}
}

func TestTrainingGoalDeleteHandler_RemovesGoal(t *testing.T) {
	store := &fakeTrainingGoalStore{exists: true}
	withTrainingGoalStore(t, store)

	req := httptest.NewRequest(http.MethodDelete, "/api/training-goal", nil)
	w := httptest.NewRecorder()

	trainingGoalDeleteHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected status 204 No Content, got %d", resp.StatusCode)
	}
	if store.deleteCount != 1 {
		t.Fatalf("Expected one delete, got %d", store.deleteCount)
	}
	if store.exists {
		t.Fatal("Expected goal to be removed")
	}
}

func TestTrainingGoalDeleteHandler_IsIdempotentWhenNoGoalExists(t *testing.T) {
	store := &fakeTrainingGoalStore{}
	withTrainingGoalStore(t, store)

	req := httptest.NewRequest(http.MethodDelete, "/api/training-goal", nil)
	w := httptest.NewRecorder()

	trainingGoalDeleteHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected status 204 No Content, got %d", resp.StatusCode)
	}
}

func TestTrainingGoalDeleteHandler_ReturnsInternalServerErrorWhenStoreFails(t *testing.T) {
	store := &fakeTrainingGoalStore{deleteErr: errors.New("database unavailable")}
	withTrainingGoalStore(t, store)

	req := httptest.NewRequest(http.MethodDelete, "/api/training-goal", nil)
	w := httptest.NewRecorder()

	trainingGoalDeleteHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected status 500 Internal Server Error, got %d", resp.StatusCode)
	}
}
