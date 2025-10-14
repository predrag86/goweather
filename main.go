package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"goweather/internal/api"
	appLog "goweather/internal/log"
	"goweather/internal/ui"
)

func main() {
	city := flag.String("city", "", "City name (e.g., belgrade)")
	hourly := flag.Bool("hourly", false, "Show hourly forecast")
	hours := flag.Int("hours", 6, "Number of hours to show (0 = all)")
	color := flag.String("color", "auto", "Color theme: auto, dark, light, none")
	emoji := flag.String("emoji", "auto", "Emoji display: auto, on, off")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	appLog.Init(*verbose)
	defer appLog.Sync()

	appLog.Info("Starting GoWeather CLI...")
	if *city == "" {
		fmt.Println("Usage: goweather -city=belgrade [--hourly] [--hours=N] [--verbose]")
		os.Exit(1)
	}

	theme := ui.GetTheme(*color, *emoji)
	appLog.Info("Selected city: %s", *city)

	coords, err := api.GetCoordinates(*city)
	if err != nil {
		appLog.Fatal("Geocoding failed: %v", err)
	}
	appLog.Info("Coordinates resolved: %.2f, %.2f", coords.Latitude, coords.Longitude)

	if *hourly {
		printHourly(coords, theme, *hours)
		return
	}

	printCurrent(coords, theme)
	appLog.Info("Execution complete.")
}

func printHourly(coords *api.Coordinates, theme ui.Theme, hours int) {
	forecast, err := api.GetHourly(coords.Latitude, coords.Longitude)
	if err != nil {
		log.Fatalf("%sHourly fetch failed:%s %v\n", theme.Red, theme.Reset, err)
	}

	fmt.Printf("\n%sHourly forecast for %s%s", theme.Bold, coords.Name, theme.Reset)
	if hours > 0 {
		fmt.Printf(" (next %d h):\n\n", hours)
	} else {
		fmt.Printf(" (full day):\n\n")
	}

	// Initialize tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Table header
	fmt.Fprintf(w, "%s%-20s\t%10s\t%10s\t%12s\t%14s\t%-18s%s\n",
		theme.Bold, "Time", "Temp °C", "Wind km/h", "Humidity %", "Pressure hPa", "Conditions", theme.Reset)
	fmt.Fprintf(w, "%s───────────────────────\t────────────\t────────────\t──────────────\t────────────\t──────────────────%s\n",
		theme.Gray, theme.Reset)

	// Limit how many rows we print
	total := len(forecast.Time())
	if hours > 0 && hours < total {
		total = hours
	}

	// Safety check: ensure slices have equal length
	for i := 0; i < total &&
		i < len(forecast.Temperature()) &&
		i < len(forecast.Windspeed()) &&
		i < len(forecast.Humidity()) &&
		i < len(forecast.Pressure()) &&
		i < len(forecast.Weathercode()); i++ {

		cond := api.WeatherDescription(forecast.Weathercode()[i])
		if !theme.Emoji {
			cond = stripEmojis(cond)
		}

		fmt.Fprintf(w, "%s%-20s%s\t%s%12.1f%s\t%s%12.1f%s\t%s%15.0f%s\t%s%20.0f%s\t%s%-20s%s\n",
			theme.Gray, forecast.Time()[i], theme.Reset,
			theme.Cyan, forecast.Temperature()[i], theme.Reset,
			theme.Yellow, forecast.Windspeed()[i], theme.Reset,
			theme.Blue, forecast.Humidity()[i], theme.Reset,
			theme.Cyan, forecast.Pressure()[i], theme.Reset,
			theme.Green, cond, theme.Reset,
		)

	}

	w.Flush()
	fmt.Println()
}

func printCurrent(coords *api.Coordinates, theme ui.Theme) {
	weather, err := api.GetWeather(coords.Latitude, coords.Longitude)
	if err != nil {
		log.Fatalf("%sWeather fetch failed:%s %v\n", theme.Red, theme.Reset, err)
	}

	desc := api.WeatherDescription(weather.Current.Weathercode)
	if !theme.Emoji {
		desc = stripEmojis(desc)
	}

	icon := ""
	if theme.Emoji {
		icon = "☁️ "
	}

	fmt.Printf("\n%sWeather in %s%s\n", theme.Bold, coords.Name, theme.Reset)
	fmt.Println("───────────────────────────────")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "%s%sTemperature:%s\t%s%.1f °C%s\n", theme.Bold, icon, theme.Reset, theme.Cyan, weather.Current.Temperature, theme.Reset)
	fmt.Fprintf(w, "%sWind:%s\t%s%.1f km/h%s %s(dir %d°)%s\n", theme.Bold, theme.Reset, theme.Yellow, weather.Current.Windspeed, theme.Reset, theme.Gray, weather.Current.Winddirection, theme.Reset)
	fmt.Fprintf(w, "%sConditions:%s\t%s%s%s\n", theme.Bold, theme.Reset, theme.Green, desc, theme.Reset)
	fmt.Fprintf(w, "%sTime:%s\t%s%s%s\n", theme.Bold, theme.Reset, theme.Blue, weather.Current.Time, theme.Reset)
	fmt.Fprintf(w, "%sPressure:%s\t%s%.0f hPa%s\n", theme.Bold, theme.Reset, theme.Cyan, weather.Current.Pressure, theme.Reset)
	fmt.Fprintf(w, "%sHumidity:%s\t%s%.0f%%%s\n", theme.Bold, theme.Reset, theme.Blue, weather.Current.Humidity, theme.Reset)

	w.Flush()
	fmt.Println()
}

func stripEmojis(s string) string {
	runes := []rune{}
	for _, r := range s {
		if r > 127 { // simple heuristic to skip emoji
			continue
		}
		runes = append(runes, r)
	}
	return string(runes)
}
