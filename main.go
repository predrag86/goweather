package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"os"
	"sync"
	"text/tabwriter"
	"time"

	"goweather/internal/api"
	"goweather/internal/cache"
	"goweather/internal/config"
	"goweather/internal/log"
	"goweather/internal/model"
	"goweather/internal/ui"
)

func main() {
	// --- Load configuration ---
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	// Register model types for gob cache
	gob.Register(&model.WeatherResponse{})
	gob.Register(&model.HourlyForecast{})

	// --- CLI overrides (highest priority) ---
	city := flag.String("city", cfg.City, "City name (e.g., belgrade)")
	hours := flag.Int("hours", cfg.Hours, "Number of hours for forecast (0=all)")
	emoji := flag.Bool("emoji", cfg.Emoji, "Enable emoji")
	color := flag.String("color", cfg.Color, "Color theme: auto|dark|light|none")
	verbose := flag.Bool("verbose", cfg.Verbose, "Verbose logging")
	mode := flag.String("mode", cfg.ForecastMode, "Forecast mode: current|hourly|both")
	flag.Parse()

	// --- Initialize file-only logger ---
	log.Init(*verbose)
	defer log.Sync()

	// --- Initialize theme ---
	emojiMode := "off"
	if *emoji {
		emojiMode = "on"
	}
	theme := ui.GetTheme(*color, emojiMode)

	// --- Initialize cache (configurable expiry) ---
	c := cache.NewCache(cfg.CacheDuration)

	// --- Get coordinates ---
	coords, err := api.GetCoordinates(*city)
	if err != nil {
		log.Logger.Fatalw("Geocoding failed", "error", err)
	}

	// --- Define cache key ---
	key := fmt.Sprintf("%s_%s", *city, *mode)

	// --- Try cache first ---
	if data, ok := c.Get(key); ok && *mode != "both" {
		log.Logger.Infow("Cache hit", "city", *city, "mode", *mode)
		switch *mode {
		case "hourly":
			printHourly(data.(*model.HourlyForecast), theme, *hours, cfg)
		default:
			printCurrent(data.(*model.WeatherResponse), theme)
		}
		return
	}

	// --- Handle all forecast modes ---
	switch *mode {
	case "hourly":
		result, err := api.GetHourly(coords.Latitude, coords.Longitude)
		if err != nil {
			log.Logger.Fatalw("Hourly fetch failed", "error", err)
		}
		c.Set(key, result)
		c.BackgroundRefresh(key, func() (any, error) {
			return api.GetHourly(coords.Latitude, coords.Longitude)
		})
		printHourly(result, theme, *hours, cfg)

	case "both":
		runBothMode(coords, c, city, hours, theme, cfg)

	default:
		result, err := api.GetWeather(coords.Latitude, coords.Longitude)
		if err != nil {
			log.Logger.Fatalw("Current fetch failed", "error", err)
		}
		c.Set(key, result)
		c.BackgroundRefresh(key, func() (any, error) {
			return api.GetWeather(coords.Latitude, coords.Longitude)
		})
		printCurrent(result, theme)
	}
}

// ---------------------------------------------------------------
// Mode “both”: concurrent fetching of current + hourly forecasts
// ---------------------------------------------------------------
func runBothMode(coords *api.Coordinates, c *cache.Cache, city *string, hours *int, theme ui.Theme, cfg *config.Config) {
	curKey := fmt.Sprintf("%s_current", *city)
	hrsKey := fmt.Sprintf("%s_hourly", *city)

	curCh := make(chan *model.WeatherResponse, 1)
	hrsCh := make(chan *model.HourlyForecast, 1)
	errCh := make(chan error, 2)

	// Try cache first
	if data, ok := c.Get(curKey); ok {
		curCh <- data.(*model.WeatherResponse)
		log.Logger.Debugw("Cache hit for current", "city", *city)
	}
	if data, ok := c.Get(hrsKey); ok {
		hrsCh <- data.(*model.HourlyForecast)
		log.Logger.Debugw("Cache hit for hourly", "city", *city)
	}

	// Concurrent refresh
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		data, err := api.GetWeather(coords.Latitude, coords.Longitude)
		if err != nil {
			errCh <- fmt.Errorf("current: %v", err)
			return
		}
		c.Set(curKey, data)
		curCh <- data
		log.Logger.Debugw("Fetched current weather", "city", *city)
	}()

	go func() {
		defer wg.Done()
		data, err := api.GetHourly(coords.Latitude, coords.Longitude)
		if err != nil {
			errCh <- fmt.Errorf("hourly: %v", err)
			return
		}
		c.Set(hrsKey, data)
		hrsCh <- data
		log.Logger.Debugw("Fetched hourly weather", "city", *city)
	}()

	// Wait for both results
	var currentData *model.WeatherResponse
	var hourlyData *model.HourlyForecast
	timeout := time.After(10 * time.Second)

	for currentData == nil || hourlyData == nil {
		select {
		case cur := <-curCh:
			currentData = cur
		case hrs := <-hrsCh:
			hourlyData = hrs
		case err := <-errCh:
			log.Logger.Errorw("Fetch failed", "error", err)
		case <-timeout:
			log.Logger.Warn("Timeout waiting for both forecasts")
			break
		}
		if currentData != nil && hourlyData != nil {
			break
		}
	}

	// Print both
	if currentData != nil {
		printCurrent(currentData, theme)
	}
	if hourlyData != nil {
		printHourly(hourlyData, theme, *hours, cfg)
	} else {
		fmt.Println("Hourly forecast unavailable (timeout or error).")
	}

	// Background refresh
	c.BackgroundRefresh(curKey, func() (any, error) {
		return api.GetWeather(coords.Latitude, coords.Longitude)
	})
	c.BackgroundRefresh(hrsKey, func() (any, error) {
		return api.GetHourly(coords.Latitude, coords.Longitude)
	})
}

// ---------------------------------------------------------------
// Helper functions for printing results
// ---------------------------------------------------------------

func printCurrent(weather *model.WeatherResponse, theme ui.Theme) {
	fmt.Printf("\n%sCurrent weather:%s\n", theme.Bold, theme.Reset)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "%s%-20s\t%-12s%s\n", theme.Bold, "Parameter", "Value", theme.Reset)
	fmt.Fprintf(w, "%s──────────────────────\t───────────────%s\n", theme.Gray, theme.Reset)

	fmt.Fprintf(w, "%sTemperature%s\t%.1f °C\n", theme.Cyan, theme.Reset, weather.Current.Temperature)
	fmt.Fprintf(w, "%sHumidity%s\t%.0f %%\n", theme.Blue, theme.Reset, weather.Current.Humidity)
	fmt.Fprintf(w, "%sWind speed%s\t%.1f km/h\n", theme.Yellow, theme.Reset, weather.Current.Windspeed)
	fmt.Fprintf(w, "%sWind direction%s\t%s\n", theme.Yellow, theme.Reset, degreesToCompass(weather.Current.Winddirection))
	fmt.Fprintf(w, "%sPressure%s\t%.0f hPa\n", theme.Green, theme.Reset, weather.Current.Pressure)
	fmt.Fprintf(w, "%sCondition%s\t%s\n", theme.Red, theme.Reset, WeatherDescription(weather.Current.Weathercode))
	w.Flush()
	fmt.Println()
}

func printHourly(forecast *model.HourlyForecast, theme ui.Theme, hours int, cfg *config.Config) {
	loc := time.Local
	locName := "Local"
	if cfg.TimeZone != "" && cfg.TimeZone != "local" {
		if userLoc, err := time.LoadLocation(cfg.TimeZone); err == nil {
			loc = userLoc
			locName = cfg.TimeZone
		}
	}

	fmt.Printf("\n%sHourly forecast (%s):%s\n", theme.Bold, locName, theme.Reset)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "%s%-20s\t%-12s\t%-12s\t%-12s\t%-12s\t%-16s\t%-16s%s\n",
		theme.Bold, "Time", "Temp (°C)", "Wind (km/h)", "Dir", "Humidity (%)", "Pressure (hPa)", "Conditions", theme.Reset)
	fmt.Fprintf(w, "%s──────────────────────\t────────────\t────────────\t────────────\t────────────\t──────────────────\t──────────────────%s\n",
		theme.Gray, theme.Reset)

	limit := len(forecast.Hourly.Time)
	if hours > 0 && hours < limit {
		limit = hours
	}

	for i := 0; i < limit; i++ {
		tStr := forecast.Hourly.Time[i]
		var tUTC time.Time
		var err error
		if len(tStr) == 16 {
			tUTC, err = time.Parse("2006-01-02T15:04", tStr)
		} else {
			tUTC, err = time.Parse(time.RFC3339, tStr)
		}
		if err != nil {
			log.Logger.Warnw("Failed to parse time", "value", tStr, "error", err)
			continue
		}
		tLocal := tUTC.In(loc)

		fmt.Fprintf(w, "%s%-20s%s\t%s%6.1f%s\t%s%6.1f%s\t%s%-4s%s\t%s%6.0f%s\t%s%6.0f%s\t%s%s%s\n",
			theme.Gray, tLocal.Format("2006-01-02 15:04"), theme.Reset,
			theme.Cyan, forecast.Hourly.Temperature[i], theme.Reset,
			theme.Yellow, forecast.Hourly.Windspeed[i], theme.Reset,
			theme.Yellow, degreesToCompass(forecast.Hourly.Winddirection[i]), theme.Reset,
			theme.Blue, forecast.Hourly.Humidity[i], theme.Reset,
			theme.Cyan, forecast.Hourly.Pressure[i], theme.Reset,
			theme.Green, WeatherDescription(forecast.Hourly.Weathercode[i]), theme.Reset)
	}
	w.Flush()
	fmt.Println()
}

// ---------------------------------------------------------------
// Utility helpers
// ---------------------------------------------------------------

func degreesToCompass(deg float64) string {
	dirs := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	idx := int((deg+22.5)/45.0) % 8
	return dirs[idx]
}

func stripEmojis(s string) string {
	runes := []rune{}
	for _, r := range s {
		if r <= 127 {
			runes = append(runes, r)
		}
	}
	return string(runes)
}

func WeatherDescription(code int) string {
	switch code {
	case 0:
		return "☀️ Clear sky"
	case 1, 2:
		return "🌤️ Partly cloudy"
	case 3:
		return "☁️ Overcast"
	case 45, 48:
		return "🌫️ Fog"
	case 51, 53, 55:
		return "🌦️ Drizzle"
	case 61, 63, 65:
		return "🌧️ Rain"
	case 80, 81, 82:
		return "🌧️ Rain showers"
	case 95:
		return "⛈️ Thunderstorm"
	case 96, 99:
		return "🌩️ Thunderstorm with hail"
	default:
		return "🌈 Unknown"
	}
}
