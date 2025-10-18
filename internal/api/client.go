package api

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"goweather/internal/log"
	"goweather/internal/model"
)

// doWithRetry performs HTTP GET with exponential backoff.
func doWithRetry(url string, maxRetries int) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 0; i < maxRetries; i++ {
		resp, err = http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}
		delay := time.Duration(math.Pow(2, float64(i))) * time.Second
		log.Logger.Warnw("Request failed, retrying",
			"url", url, "attempt", i+1, "wait", delay)
		time.Sleep(delay)
	}
	return resp, fmt.Errorf("all retries failed: %v", err)
}

// GetWeather fetches current weather data with retry/backoff.
func GetWeather(lat, lon float64) (*model.WeatherResponse, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m,windspeed_10m,winddirection_10m,weathercode,surface_pressure",
		lat, lon)

	log.Logger.Infow("Requesting current weather", "lat", lat, "lon", lon)

	resp, err := doWithRetry(url, 3)
	if err != nil {
		log.Logger.Errorw("HTTP request failed after retries", "url", url, "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	var w model.WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		log.Logger.Errorw("JSON decode failed", "error", err)
		return nil, fmt.Errorf("decode error: %v", err)
	}

	log.Logger.Infow("Weather data retrieved",
		"temperature", w.Current.Temperature,
		"wind", w.Current.Windspeed,
		"pressure", w.Current.Pressure,
		"humidity", w.Current.Humidity)

	return &w, nil
}

// GetHourly fetches hourly forecast data with retry/backoff.
func GetHourly(lat, lon float64) (*model.HourlyForecast, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&hourly=temperature_2m,relative_humidity_2m,windspeed_10m,winddirection_10m,weathercode,surface_pressure&forecast_days=1",
		lat, lon)

	log.Logger.Infow("Requesting hourly forecast", "lat", lat, "lon", lon)

	resp, err := doWithRetry(url, 3)
	if err != nil {
		log.Logger.Errorw("HTTP request failed after retries", "url", url, "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	var h model.HourlyForecast
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		log.Logger.Errorw("JSON decode failed", "error", err)
		return nil, fmt.Errorf("decode error: %v", err)
	}

	log.Logger.Infow("Hourly data retrieved", "records", len(h.Hourly.Time))
	return &h, nil
}
