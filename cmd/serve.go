package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/current", func(w http.ResponseWriter, r *http.Request) {
		handleCurrent(w, r, c)
	})
	mux.HandleFunc("/api/v1/hourly", func(w http.ResponseWriter, r *http.Request) {
		handleHourly(w, r, c)
	})

	addr := fmt.Sprintf(":%d", port)

	// ⭐ ADD MIDDLEWARE HERE (Step 2)
	srv := &http.Server{
		Addr:    addr,
		Handler: loggingMiddleware(mux), // ← Here!
	}

	// Start server in goroutine
	go func() {
		log.Logger.Infow("Starting HTTP server", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Logger.Fatalw("Server failed", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Logger.Infow("Shutdown signal received, shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Logger.Errorw("Server forced to shutdown", "error", err)
	} else {
		log.Logger.Infow("Server stopped gracefully")
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

// loggingMiddleware logs every HTTP request using zap.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Wrap ResponseWriter so we can capture status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		start := time.Now()
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)

		log.Logger.Infow("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", lrw.statusCode,
			"duration_ms", duration.Milliseconds(),
			"client_ip", r.RemoteAddr,
		)
	})
}

// loggingResponseWriter allows us to capture response status
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
