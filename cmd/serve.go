package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"goweather/internal/api"
	"goweather/internal/cache"
	"goweather/internal/config"
	"goweather/internal/log"
	"goweather/internal/model"

	"github.com/spf13/cobra"
)

var port int

func init() {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start a lightweight HTTP API service",
		Long: `Starts a local web service that exposes cached weather data as JSON.
Examples:
  goweather serve --port 8080
Then open:
  http://localhost:8080/api/v1/current?city=belgrade
  http://localhost:8080/api/v1/hourly?city=belgrade&hours=6`,
		Run: runServer,
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port number to run the server on")
	rootCmd.AddCommand(cmd)
}

func runServer(cmd *cobra.Command, args []string) {
	cfg, _ := config.Load()
	log.Init(cfg.Verbose)
	defer log.Sync()

	c := cache.NewCache(cfg.CacheDuration)

	http.HandleFunc("/api/v1/current", func(w http.ResponseWriter, r *http.Request) {
		handleCurrent(w, r, c)
	})
	http.HandleFunc("/api/v1/hourly", func(w http.ResponseWriter, r *http.Request) {
		handleHourly(w, r, c)
	})

	addr := fmt.Sprintf(":%d", port)
	log.Logger.Infow("Starting HTTP server", "port", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Logger.Fatalw("Server failed", "error", err)
	}
}

func handleCurrent(w http.ResponseWriter, r *http.Request, c *cache.Cache) {
	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, "Missing 'city' parameter", http.StatusBadRequest)
		return
	}
	key := fmt.Sprintf("%s_current", city)

	if data, ok := c.Get(key); ok {
		writeJSON(w, data.(*model.WeatherResponse))
		return
	}

	coords, err := api.GetCoordinates(city)
	if err != nil {
		http.Error(w, "Geocoding failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	res, err := api.GetWeather(coords.Latitude, coords.Longitude)
	if err != nil {
		http.Error(w, "Fetch failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	c.Set(key, res)
	writeJSON(w, res)
}

func handleHourly(w http.ResponseWriter, r *http.Request, c *cache.Cache) {
	city := r.URL.Query().Get("city")
	hoursStr := r.URL.Query().Get("hours")
	if city == "" {
		http.Error(w, "Missing 'city' parameter", http.StatusBadRequest)
		return
	}

	// Default to 6 hours if not provided
	hours := 6
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil {
			hours = h
		}
	}

	key := fmt.Sprintf("%s_hourly", city)
	if data, ok := c.Get(key); ok {
		forecast := data.(*model.HourlyForecast)
		writeLimitedHourlyJSON(w, forecast, hours)
		return
	}

	coords, err := api.GetCoordinates(city)
	if err != nil {
		http.Error(w, "Geocoding failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := api.GetHourly(coords.Latitude, coords.Longitude)
	if err != nil {
		http.Error(w, "Fetch failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	c.Set(key, res)
	writeLimitedHourlyJSON(w, res, hours)
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func writeLimitedHourlyJSON(w http.ResponseWriter, forecast *model.HourlyForecast, hours int) {
	limit := len(forecast.Hourly.Time)
	if hours > 0 && hours < limit {
		forecast.Hourly.Time = forecast.Hourly.Time[:hours]
		forecast.Hourly.Temperature = forecast.Hourly.Temperature[:hours]
		forecast.Hourly.Windspeed = forecast.Hourly.Windspeed[:hours]
		forecast.Hourly.Humidity = forecast.Hourly.Humidity[:hours]
		forecast.Hourly.Pressure = forecast.Hourly.Pressure[:hours]
		forecast.Hourly.Winddirection = forecast.Hourly.Winddirection[:hours]
		forecast.Hourly.Weathercode = forecast.Hourly.Weathercode[:hours]
	}

	writeJSON(w, forecast)
}
