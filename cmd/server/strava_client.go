package main

import (
	"net/http"
	"time"
)

const stravaAPITimeout = 10 * time.Second

var stravaHTTPClient = &http.Client{Timeout: stravaAPITimeout}
var stravaAPIBaseURL = "https://www.strava.com/api/v3"
