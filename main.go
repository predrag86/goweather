package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"os"
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
	mode := flag.String("mode", cfg.ForecastMode, "Forecast mode: current|hourly")
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
	if data, ok := c.Get(key); ok {
		log.Logger.Infow("Cache hit", "city", *city, "mode", *mode)
		switch *mode {
		case "hourly":
			printHourly(data.(*model.HourlyForecast), theme, *hours, cfg)
		default:
			printCurrent(data.(*model.WeatherResponse), theme)
		}
		return
	}

	// --- Otherwise fetch new data + store + refresh in background ---
	if *mode == "hourly" {
		result, err := api.GetHourly(coords.Latitude, coords.Longitude)
		if err != nil {
			log.Logger.Fatalw("Hourly fetch failed", "error", err)
		}
		c.Set(key, result)
		c.BackgroundRefresh(key, func() (any, error) {
			return api.GetHourly(coords.Latitude, coords.Longitude)
		})
		printHourly(result, theme, *hours, cfg)
	} else {
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
// Helper functions for printing results
// ---------------------------------------------------------------

// printCurrent prints the current weather summary.
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

// printHourly prints the hourly forecast for the next N hours (in chosen time zone).
func printHourly(forecast *model.HourlyForecast, theme ui.Theme, hours int, cfg *config.Config) {
	// Determine time zone (from config or system local)
	loc := time.Local
	locName := "Local"
	if cfg.TimeZone != "" && cfg.TimeZone != "local" {
		userLoc, err := time.LoadLocation(cfg.TimeZone)
		if err == nil {
			loc = userLoc
			locName = cfg.TimeZone
		} else {
			fmt.Printf("%sWarning:%s invalid time zone '%s', using local time\n",
				theme.Yellow, theme.Reset, cfg.TimeZone)
		}
	}

	fmt.Printf("\n%sHourly forecast (%s):%s\n", theme.Bold, locName, theme.Reset)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "%s%-20s\t%-12s\t%-12s\t%-12s\t%-12s\t%-12s\t%-16s%s\n",
		theme.Bold, "Time", "Temp (°C)", "Wind (km/h)", "Dir (°)", "Humidity (%)", "Pressure (hPa)", "Conditions", theme.Reset)
	fmt.Fprintf(w, "%s──────────────────────\t────────────\t────────────\t────────────\t────────────\t────────────\t──────────────────%s\n",
		theme.Gray, theme.Reset)

	limit := len(forecast.Hourly.Time)
	if hours > 0 && hours < limit {
		limit = hours
	}

	for i := 0; i < limit; i++ {
		// Parse UTC time and convert to chosen time zone
		tUTC, err := time.Parse(time.RFC3339, forecast.Hourly.Time[i])
		if err != nil {
			tUTC, _ = time.Parse("2006-01-02T15:04", forecast.Hourly.Time[i])
		}
		tLocal := tUTC.In(loc)

		cond := WeatherDescription(forecast.Hourly.Weathercode[i])
		if !theme.Emoji {
			cond = stripEmojis(cond)
		}

		fmt.Fprintf(w, "%s%-20s%s\t%s%6.1f%s\t%s%6.1f%s\t%s%6.0f° %s%-3s%s\t%s%6.0f%s\t%s%6.0f%s\t%s%-16s%s\n",
			theme.Gray, tLocal.Format("2006-01-02 15:04"), theme.Reset,
			theme.Cyan, forecast.Hourly.Temperature[i], theme.Reset,
			theme.Yellow, forecast.Hourly.Windspeed[i], theme.Reset,
			theme.Yellow, forecast.Hourly.Winddirection[i], theme.Gray, degreesToCompass(forecast.Hourly.Winddirection[i]), theme.Reset,
			theme.Blue, forecast.Hourly.Humidity[i], theme.Reset,
			theme.Cyan, forecast.Hourly.Pressure[i], theme.Reset,
			theme.Green, cond, theme.Reset)

	}

	w.Flush()
	fmt.Println()
}

// stripEmojis removes emojis when user disables them.
func stripEmojis(s string) string {
	runes := []rune{}
	for _, r := range s {
		if r <= 127 { // basic ASCII
			runes = append(runes, r)
		}
	}
	return string(runes)
}

func degreesToCompass(deg float64) string {
	dirs := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	idx := int((deg+22.5)/45.0) % 8
	return dirs[idx]
}

// WeatherDescription converts weather codes to text.
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
