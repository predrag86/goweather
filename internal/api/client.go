package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"goweather/internal/log"
	"goweather/internal/model"
)

// GetWeather fetches current weather data for given coordinates.
// Includes temperature, wind, weathercode, humidity, and pressure.
func GetWeather(lat, lon float64) (*model.WeatherResponse, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m,windspeed_10m,winddirection_10m,weathercode,surface_pressure",
		lat, lon)

	log.Logger.Infow("Requesting current weather",
		"lat", lat,
		"lon", lon,
		"url", url,
	)

	resp, err := http.Get(url)
	if err != nil {
		log.Logger.Errorw("HTTP request failed", "url", url, "error", err)
		return nil, fmt.Errorf("weather request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Logger.Errorw("Non-200 status from weather API",
			"status", resp.Status, "url", url)
		return nil, fmt.Errorf("weather API failed: %s", resp.Status)
	}

	var w model.WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		log.Logger.Errorw("JSON decode failed", "error", err)
		return nil, fmt.Errorf("decode error: %v", err)
	}

	log.Logger.Infow("Weather data retrieved",
		"temperature", w.Current.Temperature,
		"wind", w.Current.Windspeed,
		"pressure", w.Current.Pressure,
		"humidity", w.Current.Humidity,
	)

	return &w, nil
}

// GetHourly fetches hourly forecast data for the given coordinates.
// Includes temperature, humidity, windspeed, pressure, and weathercode.
func GetHourly(lat, lon float64) (*model.HourlyForecast, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&hourly=temperature_2m,relative_humidity_2m,windspeed_10m,weathercode,surface_pressure&forecast_days=1",
		lat, lon)

	log.Logger.Infow("Requesting hourly forecast",
		"lat", lat,
		"lon", lon,
		"url", url,
	)

	resp, err := http.Get(url)
	if err != nil {
		log.Logger.Errorw("HTTP request failed", "url", url, "error", err)
		return nil, fmt.Errorf("hourly request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Logger.Errorw("Non-200 status from hourly API",
			"status", resp.Status, "url", url)
		return nil, fmt.Errorf("hourly API failed: %s", resp.Status)
	}

	var h model.HourlyForecast
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		log.Logger.Errorw("JSON decode failed", "error", err)
		return nil, fmt.Errorf("decode error: %v", err)
	}

	log.Logger.Infow("Hourly data retrieved",
		"records", len(h.Hourly.Time),
	)

	return &h, nil
}
