package main

import (
	"strings"
	"testing"
	"time"
)

func TestMapStravaActivities_MapsCanonicalFields(t *testing.T) {
	elevationGain := 51.2
	averageHeartRate := 145.5
	maxHeartRate := 181.0
	calories := 654.3
	startedAt := time.Date(2026, 5, 15, 8, 30, 0, 0, time.FixedZone("CEST", 2*60*60))

	activities, err := mapStravaActivities("00000000-0000-0000-0000-000000000042", []StravaActivity{
		{
			ID:                 123,
			Type:               "Run",
			SportType:          "TrailRun",
			Name:               "Forest Loop",
			StartDate:          startedAt,
			ElapsedTime:        3600,
			MovingTime:         3500,
			Distance:           10001.5,
			TotalElevationGain: &elevationGain,
			AverageHeartRate:   &averageHeartRate,
			MaxHeartRate:       &maxHeartRate,
			Calories:           &calories,
		},
	}, "raw/object.json")
	if err != nil {
		t.Fatalf("Expected mapped activity, got error: %v", err)
	}

	if len(activities) != 1 {
		t.Fatalf("Expected one mapped activity, got %d", len(activities))
	}

	activity := activities[0]
	if activity.UserID != "00000000-0000-0000-0000-000000000042" {
		t.Fatalf("Expected user id to be mapped, got %q", activity.UserID)
	}
	if activity.Provider != "strava" {
		t.Fatalf("Expected provider strava, got %q", activity.Provider)
	}
	if activity.ProviderID != "123" {
		t.Fatalf("Expected provider id 123, got %q", activity.ProviderID)
	}
	if activity.ActivityType != "run" {
		t.Fatalf("Expected normalized activity type run, got %q", activity.ActivityType)
	}
	if activity.Name != "Forest Loop" {
		t.Fatalf("Expected name Forest Loop, got %q", activity.Name)
	}
	if !activity.StartedAt.Equal(startedAt.UTC()) {
		t.Fatalf("Expected UTC start time %s, got %s", startedAt.UTC(), activity.StartedAt)
	}
	if activity.DurationSeconds != 3600 {
		t.Fatalf("Expected duration 3600, got %d", activity.DurationSeconds)
	}
	if activity.MovingSeconds != 3500 {
		t.Fatalf("Expected moving seconds 3500, got %d", activity.MovingSeconds)
	}
	if activity.DistanceMeters != 10001.5 {
		t.Fatalf("Expected distance 10001.5, got %f", activity.DistanceMeters)
	}
	if activity.ElevationGainMeters == nil || *activity.ElevationGainMeters != elevationGain {
		t.Fatalf("Expected elevation gain %f, got %v", elevationGain, activity.ElevationGainMeters)
	}
	if activity.AverageHeartRateBPM == nil || *activity.AverageHeartRateBPM != averageHeartRate {
		t.Fatalf("Expected average heart rate %f, got %v", averageHeartRate, activity.AverageHeartRateBPM)
	}
	if activity.MaxHeartRateBPM == nil || *activity.MaxHeartRateBPM != maxHeartRate {
		t.Fatalf("Expected max heart rate %f, got %v", maxHeartRate, activity.MaxHeartRateBPM)
	}
	if activity.CaloriesKcal == nil || *activity.CaloriesKcal != calories {
		t.Fatalf("Expected calories %f, got %v", calories, activity.CaloriesKcal)
	}
	if activity.RawObjectKey != "raw/object.json" {
		t.Fatalf("Expected raw object key, got %q", activity.RawObjectKey)
	}
}

func TestMapStravaActivities_AllowsOptionalFieldsToBeMissing(t *testing.T) {
	activities, err := mapStravaActivities("00000000-0000-0000-0000-000000000042", []StravaActivity{
		{
			ID:          123,
			Type:        "Run",
			Name:        "Easy Run",
			StartDate:   time.Date(2026, 5, 15, 6, 30, 0, 0, time.UTC),
			ElapsedTime: 1800,
			MovingTime:  1700,
			Distance:    5000,
		},
	}, "raw/object.json")
	if err != nil {
		t.Fatalf("Expected mapped activity, got error: %v", err)
	}

	activity := activities[0]
	if activity.ElevationGainMeters != nil {
		t.Fatalf("Expected nil elevation gain, got %v", activity.ElevationGainMeters)
	}
	if activity.AverageHeartRateBPM != nil {
		t.Fatalf("Expected nil average heart rate, got %v", activity.AverageHeartRateBPM)
	}
	if activity.MaxHeartRateBPM != nil {
		t.Fatalf("Expected nil max heart rate, got %v", activity.MaxHeartRateBPM)
	}
	if activity.CaloriesKcal != nil {
		t.Fatalf("Expected nil calories, got %v", activity.CaloriesKcal)
	}
}

func TestMapStravaActivities_RequiresCoreFields(t *testing.T) {
	tests := []struct {
		name     string
		activity StravaActivity
		wantErr  string
	}{
		{
			name: "missing id",
			activity: StravaActivity{
				Name:      "Run",
				StartDate: time.Date(2026, 5, 15, 6, 30, 0, 0, time.UTC),
			},
			wantErr: "id is required",
		},
		{
			name: "missing name",
			activity: StravaActivity{
				ID:        123,
				StartDate: time.Date(2026, 5, 15, 6, 30, 0, 0, time.UTC),
			},
			wantErr: "name is required",
		},
		{
			name: "missing start date",
			activity: StravaActivity{
				ID:   123,
				Name: "Run",
			},
			wantErr: "start_date is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mapStravaActivities("00000000-0000-0000-0000-000000000042", []StravaActivity{tt.activity}, "raw/object.json")
			if err == nil {
				t.Fatal("Expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestNormalizeStravaActivityType(t *testing.T) {
	tests := []struct {
		name     string
		activity StravaActivity
		want     string
	}{
		{name: "sport type wins", activity: StravaActivity{Type: "Ride", SportType: "Run"}, want: "run"},
		{name: "run variants", activity: StravaActivity{SportType: "VirtualRun"}, want: "run"},
		{name: "ride variants", activity: StravaActivity{SportType: "GravelRide"}, want: "ride"},
		{name: "swim", activity: StravaActivity{SportType: "Swim"}, want: "swim"},
		{name: "walk variants", activity: StravaActivity{SportType: "Hike"}, want: "walk"},
		{name: "workout variants", activity: StravaActivity{SportType: "WeightTraining"}, want: "workout"},
		{name: "yoga", activity: StravaActivity{SportType: "Yoga"}, want: "yoga"},
		{name: "fallback trims and lowercases", activity: StravaActivity{SportType: " AlpineSki "}, want: "alpineski"},
		{name: "empty fallback", activity: StravaActivity{}, want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeStravaActivityType(tt.activity)
			if got != tt.want {
				t.Fatalf("Expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestStravaActivitiesObjectKey(t *testing.T) {
	fetchedAt := time.Date(2026, 5, 15, 8, 30, 1, 123456789, time.FixedZone("CEST", 2*60*60))

	got := stravaActivitiesObjectKey("00000000-0000-0000-0000-000000000042", fetchedAt)
	want := "strava/users/00000000-0000-0000-0000-000000000042/activity-lists/20260515T063001.123456789Z.json"
	if got != want {
		t.Fatalf("Expected object key %q, got %q", want, got)
	}
}
