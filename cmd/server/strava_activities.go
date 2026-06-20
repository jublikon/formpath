package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const stravaActivitiesPerPage = 200

type StravaActivity struct {
	ID                 int64     `json:"id"`
	Type               string    `json:"type"`
	SportType          string    `json:"sport_type"`
	Name               string    `json:"name"`
	StartDate          time.Time `json:"start_date"`
	ElapsedTime        int       `json:"elapsed_time"`
	MovingTime         int       `json:"moving_time"`
	Distance           float64   `json:"distance"`
	TotalElevationGain *float64  `json:"total_elevation_gain"`
	AverageHeartRate   *float64  `json:"average_heartrate"`
	MaxHeartRate       *float64  `json:"max_heartrate"`
	Calories           *float64  `json:"calories"`
}

type HTTPStatusError struct {
	code int
	err  error
}

func (e HTTPStatusError) Error() string {
	return fmt.Sprintf("HTTP %d: %v", e.code, e.err)
}

func (e HTTPStatusError) Unwrap() error {
	return e.err
}

func fetchStravaActivitiesPayload(accessToken string) ([]byte, error) {
	if accessToken == "" {
		return nil, errors.New("access token is required")
	}

	activitiesURL, err := url.Parse(stravaAPIBaseURL + "/athlete/activities")
	if err != nil {
		return nil, fmt.Errorf("parsing Strava activities URL: %w", err)
	}

	query := activitiesURL.Query()
	query.Set("page", "1")
	query.Set("per_page", strconv.Itoa(stravaActivitiesPerPage))
	activitiesURL.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, activitiesURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating Strava activities request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := stravaHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling Strava activities endpoint: %w", err)
	}

	body, err := readAllAndClose(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading Strava activities response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, HTTPStatusError{
			code: resp.StatusCode,
			err:  fmt.Errorf("Strava activities request failed with status %d", resp.StatusCode),
		}
	}

	return body, nil
}

func transformStravaActivities(userID string, rawBody []byte, rawObjectKey string) ([]Activity, error) {
	var activities []StravaActivity
	if err := json.Unmarshal(rawBody, &activities); err != nil {
		return nil, fmt.Errorf("decoding raw Strava activities: %w", err)
	}

	return mapStravaActivities(userID, activities, rawObjectKey)
}

func mapStravaActivities(userID string, stravaActivities []StravaActivity, rawObjectKey string) ([]Activity, error) {
	activities := make([]Activity, 0, len(stravaActivities))
	for _, stravaActivity := range stravaActivities {
		if stravaActivity.ID == 0 {
			return nil, errors.New("Strava activity id is required")
		}
		if stravaActivity.Name == "" {
			return nil, fmt.Errorf("Strava activity %d name is required", stravaActivity.ID)
		}
		if stravaActivity.StartDate.IsZero() {
			return nil, fmt.Errorf("Strava activity %d start_date is required", stravaActivity.ID)
		}

		activities = append(activities, Activity{
			UserID:              userID,
			Provider:            "strava",
			ProviderID:          strconv.FormatInt(stravaActivity.ID, 10),
			ActivityType:        normalizeStravaActivityType(stravaActivity),
			Name:                stravaActivity.Name,
			StartedAt:           stravaActivity.StartDate.UTC(),
			DurationSeconds:     stravaActivity.ElapsedTime,
			MovingSeconds:       stravaActivity.MovingTime,
			DistanceMeters:      stravaActivity.Distance,
			ElevationGainMeters: stravaActivity.TotalElevationGain,
			AverageHeartRateBPM: stravaActivity.AverageHeartRate,
			MaxHeartRateBPM:     stravaActivity.MaxHeartRate,
			CaloriesKcal:        stravaActivity.Calories,
			RawObjectKey:        rawObjectKey,
		})
	}
	return activities, nil
}

func normalizeStravaActivityType(activity StravaActivity) string {
	activityType := activity.SportType
	if activityType == "" {
		activityType = activity.Type
	}

	switch strings.ToLower(activityType) {
	case "run", "trailrun", "virtualrun":
		return "run"
	case "ride", "mountainbikeride", "gravelride", "virtualride", "ebikeride", "emountainbikeride":
		return "ride"
	case "swim":
		return "swim"
	case "walk", "hike":
		return "walk"
	case "workout", "weighttraining", "crossfit", "elliptical", "stairstepper":
		return "workout"
	case "yoga":
		return "yoga"
	default:
		normalized := strings.ToLower(strings.TrimSpace(activityType))
		if normalized == "" {
			return "unknown"
		}
		return normalized
	}
}

func stravaActivitiesObjectKey(userID string, fetchedAt time.Time) string {
	return fmt.Sprintf(
		"strava/users/%s/activity-lists/%s.json",
		userID,
		fetchedAt.UTC().Format("20060102T150405.000000000Z"),
	)
}
