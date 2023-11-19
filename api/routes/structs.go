package routes

import "time"

type UrlShortenerRequest struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"custom_short"`
	Expiry      time.Duration `json:"expiry"`
}

type UrlShortenerResponse struct {
	URL           string        `json:"url"`
	CustomShort   string        `json:"custom_short"`
	Expiry        time.Duration `json:"expiry"`
	RateRemaining int           `json:"rate_remaining"`
	RateLimitRest time.Duration `json:"rate_limit_rest"`
}
