package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"goweather/internal/model"
)

// Current weather (now includes humidity & pressure)
func GetWeather(lat, lon float64) (*model.WeatherResponse, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m,windspeed_10m,winddirection_10m,weathercode,surface_pressure",
		lat, lon)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var w model.WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}
	return &w, nil
}

// Hourly forecast (temperature + wind + humidity + weathercode + pressure)
func GetHourly(lat, lon float64) (*model.HourlyForecast, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&hourly=temperature_2m,relative_humidity_2m,windspeed_10m,weathercode,surface_pressure&forecast_days=1",
		lat, lon)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var h model.HourlyForecast
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}
	return &h, nil
}
