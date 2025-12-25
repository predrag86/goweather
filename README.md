# 📦 GoWeather – Modern CLI & API Weather Service (Go + Cobra + Prometheus + Zap)

GoWeather is a modern, modular, high-performance **weather client and local weather API service** written in Go.

It uses the **Open-Meteo API** (no API key required), stores results in a **local cache**, supports **structured logging**, **Prometheus metrics**, **colorized output**, **emoji themes**, and exposes both a **Cobra-based CLI** and **HTTP endpoints**.

---

## 🚀 Features

### 🌤 Weather Forecasting
- Current weather  
- Hourly forecast (configurable hours)  
- Concurrent fetching (current + hourly)  
- Local timezone conversion  
- Wind direction (compass)  
- Weather code → description mapping  

### 🧰 CLI (Cobra)

```
goweather current --city belgrade
goweather hourly --city belgrade --hours 6
goweather both --city belgrade
goweather serve --port 8080
```

Includes:
- Emoji toggle (`--emoji=on|off`)
- Color theme (`--color auto|dark|light|none`)
- Verbose logging (`--verbose`)
- Config overrides via YAML

---

### 🏎 Performance
- File-based cache with expiration  
- Background cache refresh using goroutines  
- API retry/backoff  

---

### 📊 Observability
- Zap structured logging  
- Lumberjack log rotation  
- Prometheus metrics at `/metrics`  
- HTTP request logging middleware  

---

### 🌐 HTTP API Service

Start server:
```
goweather serve --port 8080
```

Endpoints:
```
GET /api/v1/current?city=belgrade
GET /api/v1/hourly?city=belgrade&hours=6
GET /metrics
```

---

## 📁 Project Structure

```
goweather/
 ├── cmd/                    # Cobra commands
 ├── internal/
 │    ├── api/               # Open-Meteo clients
 │    ├── cache/             # Time-based cache
 │    ├── cli/               # CLI rendering helpers
 │    ├── config/            # YAML config loader
 │    ├── log/               # Zap + Lumberjack logger
 │    ├── model/             # Data models
 │    └── ui/                # Themes and emojis
 ├── main.go
 ├── go.mod / go.sum
 └── README.md
```

---

## 🛠 Installation

Requires **Go 1.23+**.

```bash
git clone https://github.com/predrag86/goweather.git
cd goweather
```

Build:

```bash
go build -o goweather .
```

Run locally:

```bash
go run . <command>
```

---

## 🖥 CLI Usage

### Current Weather
```bash
goweather current --city belgrade
```

### Hourly Forecast
```bash
goweather hourly --city belgrade --hours 6
```

### Both (parallel fetch)
```bash
goweather both --city belgrade --hours 6
```

### Color & Emoji Options
```bash
goweather current --color dark --emoji off
```

---

## 🌐 Run API Server

```bash
goweather serve --port 8080
```

Current weather:
```
http://localhost:8080/api/v1/current?city=belgrade
```

Hourly forecast:
```
http://localhost:8080/api/v1/hourly?city=belgrade&hours=6
```

Prometheus metrics:
```
http://localhost:8080/metrics
```

---

## ⚙ Configuration

Configuration file location:

```
$HOME/.config/goweather/config.yaml
```

Example:

```yaml
city: "belgrade"
hours: 6
emoji: true
color: "auto"
verbose: false
timezone: "Europe/Belgrade"
cache_duration: "10m"
log_path: "$HOME/.cache/goweather/app.log"
```

CLI flags override config values.

---

## 🧪 Development & Testing

```bash
go build ./...
go test -v ./...
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

---

## 🤖 GitHub Actions

CI pipeline includes:
- Build  
- Vet  
- Formatting checks  
- Unit tests  
- Coverage summary  

---

## 📜 License

MIT (or your preferred license)
