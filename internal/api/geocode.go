package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"goweather/internal/model"
)

type Coordinates struct {
	Name      string
	Latitude  float64
	Longitude float64
}

const cacheFile = "cache.json"

type cacheMap map[string]Coordinates

// GetCoordinates returns lat/lon for a city, cached if available.
func GetCoordinates(city string) (*Coordinates, error) {
	cache, _ := loadCache()

	if val, ok := cache[city]; ok {
		return &val, nil
	}

	url := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1", city)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocode API error: %s", resp.Status)
	}

	var geo model.GeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&geo); err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}
	if len(geo.Results) == 0 {
		return nil, fmt.Errorf("no results for city: %s", city)
	}

	r := geo.Results[0]
	coord := Coordinates{Name: r.Name, Latitude: r.Latitude, Longitude: r.Longitude}

	cache[city] = coord
	saveCache(cache)

	return &coord, nil
}

func loadCache() (cacheMap, error) {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return cacheMap{}, nil // no cache yet
	}
	var c cacheMap
	_ = json.Unmarshal(data, &c)
	return c, nil
}

func saveCache(c cacheMap) {
	data, _ := json.MarshalIndent(c, "", "  ")
	_ = os.WriteFile(filepath.Clean(cacheFile), data, 0644)
}
