package cli

import (
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

// Reusable functions for CLI commands
// -----------------------------------

func RunBothMode(coords *api.Coordinates, c *cache.Cache, city *string, hours *int, theme ui.Theme, cfg *config.Config) {
	curKey := fmt.Sprintf("%s_current", *city)
	hrsKey := fmt.Sprintf("%s_hourly", *city)

	currentData, _ := c.Get(curKey)
	hourlyData, _ := c.Get(hrsKey)

	if currentData == nil {
		data, err := api.GetWeather(coords.Latitude, coords.Longitude)
		if err != nil {
			log.Logger.Fatalw("Current fetch failed", "error", err)
		}
		c.Set(curKey, data)
		currentData = data
	}

	if hourlyData == nil {
		data, err := api.GetHourly(coords.Latitude, coords.Longitude)
		if err != nil {
			log.Logger.Fatalw("Hourly fetch failed", "error", err)
		}
		c.Set(hrsKey, data)
		hourlyData = data
	}

	PrintCurrent(currentData.(*model.WeatherResponse), theme)
	PrintHourly(hourlyData.(*model.HourlyForecast), theme, *hours, cfg)

	// Background refresh for both
	c.BackgroundRefresh(curKey, func() (any, error) {
		return api.GetWeather(coords.Latitude, coords.Longitude)
	})
	c.BackgroundRefresh(hrsKey, func() (any, error) {
		return api.GetHourly(coords.Latitude, coords.Longitude)
	})
}

func PrintCurrent(weather *model.WeatherResponse, theme ui.Theme) {
	fmt.Printf("\n%sCurrent weather:%s\n", theme.Bold, theme.Reset)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "%s%-20s\t%-12s%s\n", theme.Bold, "Parameter", "Value", theme.Reset)
	fmt.Fprintf(w, "%s──────────────────────\t───────────────%s\n", theme.Gray, theme.Reset)

	fmt.Fprintf(w, "%sTemperature%s\t%.1f °C\n", theme.Cyan, theme.Reset, weather.Current.Temperature)
	fmt.Fprintf(w, "%sHumidity%s\t%.0f %%\n", theme.Blue, theme.Reset, weather.Current.Humidity)
	fmt.Fprintf(w, "%sWind speed%s\t%.1f km/h\n", theme.Yellow, theme.Reset, weather.Current.Windspeed)
	fmt.Fprintf(w, "%sWind direction%s\t%s\n", theme.Yellow, theme.Reset, degreesToCompass(weather.Current.Winddirection))
	fmt.Fprintf(w, "%sPressure%s\t%.0f hPa\n", theme.Green, theme.Reset, weather.Current.Pressure)
	fmt.Fprintf(w, "%sCondition%s\t%s\n", theme.Red, theme.Reset, api.WeatherDescription(weather.Current.Weathercode))
	w.Flush()
	fmt.Println()
}

func PrintHourly(forecast *model.HourlyForecast, theme ui.Theme, hours int, cfg *config.Config) {
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
			theme.Green, api.WeatherDescription(forecast.Hourly.Weathercode[i]), theme.Reset)
	}
	w.Flush()
	fmt.Println()
}

// Utility
func degreesToCompass(deg float64) string {
	dirs := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	idx := int((deg+22.5)/45.0) % 8
	return dirs[idx]
}
