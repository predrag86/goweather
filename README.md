📦 GoWeather – Modern CLI & API Weather Service (Go + Cobra + Prometheus + Zap)

GoWeather is a modern, modular, high-performance weather client and local weather API service written in Go.

It uses the Open-Meteo API (no API key required), stores results in a local cache, supports structured logging, Prometheus metrics, colorized output, emoji themes, and exposes both a Cobra-based CLI and HTTP endpoints.

🚀 Features
🌤 Weather Forecasting

Current weather

Hourly forecast (configurable hours)

Concurrent fetching (current + hourly)

Local timezone conversion

Wind direction (compass)

Weather code → description mapping

🧰 CLI (Cobra)
goweather current --city belgrade
goweather hourly --city belgrade --hours 6
goweather both --city belgrade
goweather serve --port 8080


Includes:

Emoji toggle (--emoji=on|off)

Color theme (--color auto|dark|light|none)

Verbose logging (--verbose)

Config overrides via YAML

🏎 Performance

Configurable file-based cache with expiration

Background cache refresh using goroutines

API retry/backoff

📊 Observability

Zap structured logging (file only)

Lumberjack log rotation

Prometheus metrics at /metrics

Request counts

Request durations

Per-method/path metrics

Request logging middleware

🌐 HTTP API Service

Run with:

goweather serve --port 8080


Endpoints:

GET /api/v1/current?city=belgrade
GET /api/v1/hourly?city=belgrade&hours=6
GET /metrics

📁 Project Structure
goweather/
 ├── cmd/                    # Cobra commands (current, hourly, both, serve)
 ├── internal/
 │    ├── api/               # Open-Meteo API clients
 │    ├── cache/             # Time-based cache with disk persistence
 │    ├── cli/               # Shared CLI printing helpers
 │    ├── config/            # YAML config loader
 │    ├── log/               # Zap + Lumberjack logger
 │    ├── model/             # Weather models
 │    └── ui/                # Color themes, emoji handling
 ├── main.go                 # Cobra entrypoint
 ├── go.mod / go.sum
 └── README.md

🛠 Installation

Requires Go 1.23+.

Clone repo:

git clone https://github.com/<your-username>/goweather.git
cd goweather


Build:

go build -o goweather .


Or run locally:

go run . <command>

🖥 CLI Usage
🌤 Current Weather
goweather current --city belgrade

🕒 Hourly Forecast
goweather hourly --city belgrade --hours 6

🔀 Both (parallel fetch)
goweather both --city belgrade --hours 6

🎨 Color & Emoji Options
goweather current --color dark --emoji off

🌐 Run API Server

Start service:

goweather serve --port 8080

Current Weather
GET http://localhost:8080/api/v1/current?city=belgrade

Hourly Forecast
GET http://localhost:8080/api/v1/hourly?city=belgrade&hours=6

Prometheus Metrics
GET http://localhost:8080/metrics

⚙ Configuration

Configuration file located at:

$HOME/.config/goweather/config.yaml


Example:

city: "belgrade"
hours: 6
emoji: true
color: "auto"
verbose: false
timezone: "Europe/Belgrade"
cache_duration: "10m"
log_path: "$HOME/.cache/goweather/app.log"


All CLI flags override config values.

🧪 Development & Testing

Run build:

go build ./...


Run tests:

go test -v ./...


Generate coverage:

go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

🤖 GitHub Actions CI

Your CI pipeline includes:

Go build

go vet

Formatting checks

Unit test execution

Coverage summary

📦 Coming Next

JSON output mode (--json)

CORS support

Grafana dashboard for metrics

Containerization with Docker

Expanded test suite

📜 License

MIT (or your chosen license)