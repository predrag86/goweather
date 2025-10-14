package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"goweather/internal/log"
	"goweather/internal/model"
)

// Coordinates holds latitude/longitude pair for a city.
type Coordinates struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Country   string  `json:"country"`
}

// cache file for geocoding lookups
const cacheFile = "geocode_cache.json"

// GetCoordinates returns coordinates for a city, using Open-Meteo’s free geocoding API.
// Results are cached to reduce network calls.
func GetCoordinates(city string) (*Coordinates, error) {
	cache, _ := loadCache()
	if val, ok := cache[city]; ok {
		log.Logger.Infow("Geocoding cache hit",
			"city", city,
			"lat", val.Latitude,
			"lon", val.Longitude,
		)
		return &val, nil
	}

	log.Logger.Infow("Calling Open-Meteo geocoding API", "city", city)
	url := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1", city)

	resp, err := http.Get(url)
	if err != nil {
		log.Logger.Errorw("HTTP request failed", "url", url, "error", err)
		return nil, fmt.Errorf("geocode request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Logger.Errorw("Geocoding API returned non-200 status",
			"status", resp.Status, "url", url)
		return nil, fmt.Errorf("geocode failed: %s", resp.Status)
	}

	var geo model.GeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&geo); err != nil {
		log.Logger.Errorw("Failed to decode geocoding response", "error", err)
		return nil, fmt.Errorf("decode failed: %v", err)
	}

	if len(geo.Results) == 0 {
		log.Logger.Warnw("No geocoding results found", "city", city)
		return nil, fmt.Errorf("no coordinates found for %s", city)
	}

	res := geo.Results[0]
	coord := Coordinates{
		Name:      res.Name,
		Latitude:  res.Latitude,
		Longitude: res.Longitude,
		Country:   res.Country,
	}

	cache[city] = coord
	_ = saveCache(cache)

	log.Logger.Infow("Geocoding success",
		"city", coord.Name,
		"lat", coord.Latitude,
		"lon", coord.Longitude,
	)

	return &coord, nil
}

// local cache helpers
func cachePath() string {
	dir, _ := os.UserCacheDir()
	path := filepath.Join(dir, "goweather", cacheFile)
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	return path
}

func loadCache() (map[string]Coordinates, error) {
	path := cachePath()
	file, err := os.Open(path)
	if err != nil {
		return make(map[string]Coordinates), nil
	}
	defer file.Close()

	var data map[string]Coordinates
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return make(map[string]Coordinates), nil
	}
	return data, nil
}

func saveCache(data map[string]Coordinates) error {
	path := cachePath()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
