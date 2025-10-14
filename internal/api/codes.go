package api

// WeatherDescription converts Open-Meteo weather codes to text/emoji.
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
	case 56, 57:
		return "🌧️ Freezing drizzle"

	case 61, 63, 65:
		return "🌧️ Rain"
	case 66, 67:
		return "🌧️ Freezing rain"

	case 71, 73, 75:
		return "🌨️ Snow"
	case 77:
		return "❄️ Snow grains"

	case 80, 81, 82:
		return "🌧️ Rain showers"
	case 85, 86:
		return "🌨️ Snow showers"

	case 95:
		return "⛈️ Thunderstorm"
	case 96, 99:
		return "🌩️ Thunderstorm with hail"

	default:
		return "🌈 Unknown"
	}
}
