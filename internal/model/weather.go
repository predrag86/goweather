package model

type WeatherResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Current   struct {
		Time          string  `json:"time"`
		Temperature   float64 `json:"temperature_2m"`
		Humidity      float64 `json:"relative_humidity_2m"`
		Windspeed     float64 `json:"windspeed_10m"`
		Winddirection float64 `json:"winddirection_10m"`
		Pressure      float64 `json:"surface_pressure"`
		Weathercode   int     `json:"weathercode"`
	} `json:"current"`
}

type HourlyForecast struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Hourly    struct {
		Time          []string  `json:"time"`
		Temperature   []float64 `json:"temperature_2m"`
		Humidity      []float64 `json:"relative_humidity_2m"`
		Windspeed     []float64 `json:"windspeed_10m"`
		Winddirection []float64 `json:"winddirection_10m"`
		Pressure      []float64 `json:"surface_pressure"`
		Weathercode   []int     `json:"weathercode"`
	} `json:"hourly"`
}

func (h *HourlyForecast) Time() []string         { return h.Hourly.Time }
func (h *HourlyForecast) Temperature() []float64 { return h.Hourly.Temperature }
func (h *HourlyForecast) Humidity() []float64    { return h.Hourly.Humidity }
func (h *HourlyForecast) Windspeed() []float64   { return h.Hourly.Windspeed }
func (h *HourlyForecast) Pressure() []float64    { return h.Hourly.Pressure }
func (h *HourlyForecast) Weathercode() []int     { return h.Hourly.Weathercode }

type GeocodeResponse struct {
	Results []struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Country   string  `json:"country"`
	} `json:"results"`
}
