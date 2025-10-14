package model

type WeatherResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Current   struct {
		Temperature   float64 `json:"temperature_2m"`
		Windspeed     float64 `json:"windspeed_10m"`
		Winddirection int     `json:"winddirection_10m"`
		Weathercode   int     `json:"weathercode"`
		Pressure      float64 `json:"surface_pressure"`
		Time          string  `json:"time"`
	} `json:"current"`
}

type HourlyForecast struct {
	Hourly struct {
		Time         []string  `json:"time"`
		Temperature2 []float64 `json:"temperature_2m"`
		Windspeed10  []float64 `json:"windspeed_10m"`
		Pressure     []float64 `json:"surface_pressure"`
		Weathercode  []int     `json:"weathercode"`
	} `json:"hourly"`
}

func (h *HourlyForecast) Time() []string         { return h.Hourly.Time }
func (h *HourlyForecast) Temperature() []float64 { return h.Hourly.Temperature2 }
func (h *HourlyForecast) Windspeed() []float64   { return h.Hourly.Windspeed10 }
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
