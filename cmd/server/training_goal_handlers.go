package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

func trainingGoalGetHandler(w http.ResponseWriter, r *http.Request) {
	cfg := loadAppConfig()

	goal, err := trainingGoalStore.GetTrainingGoal(r.Context(), cfg.AppUserID)
	if errors.Is(err, ErrTrainingGoalNotFound) {
		http.Error(w, "training goal not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to load training goal", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goal)
}

func trainingGoalPutHandler(w http.ResponseWriter, r *http.Request) {
	cfg := loadAppConfig()

	var goal TrainingGoal
	if err := json.NewDecoder(r.Body).Decode(&goal); err != nil {
		http.Error(w, "invalid training goal JSON", http.StatusBadRequest)
		return
	}

	goal.UserID = cfg.AppUserID
	goal.GoalType = trainingGoalTypeDistanceEvent

	if err := goal.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := trainingGoalStore.SaveTrainingGoal(r.Context(), goal); err != nil {
		http.Error(w, "failed to save training goal", http.StatusInternalServerError)
		return
	}

	savedGoal, err := trainingGoalStore.GetTrainingGoal(r.Context(), cfg.AppUserID)
	if err != nil {
		http.Error(w, "failed to load saved training goal", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(savedGoal)
}

func trainingGoalDeleteHandler(w http.ResponseWriter, r *http.Request) {
	cfg := loadAppConfig()

	err := trainingGoalStore.DeleteTrainingGoal(r.Context(), cfg.AppUserID)
	if err != nil && !errors.Is(err, ErrTrainingGoalNotFound) {
		http.Error(w, "failed to delete training goal", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
