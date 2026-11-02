package forecast

import "weatherwatch/internal/apperr"

var ErrForecastUnavailable = apperr.Upstream("FORECAST_UNAVAILABLE", "The weather provider could not supply the forecast")

type Forecast struct {
	Timezone string         `json:"timezone"`
	Current  CurrentWeather `json:"current"`
	Daily    []DailyWeather `json:"daily"`
}

type CurrentWeather struct {
	Time                string  `json:"time"`
	Temperature         float64 `json:"temperature"`
	ApparentTemperature float64 `json:"apparentTemperature"`
	Humidity            int     `json:"humidity"`
	WindSpeed           float64 `json:"windSpeed"`
	WeatherCode         int     `json:"weatherCode"`
	Condition           string  `json:"condition"`
	IsDay               bool    `json:"isDay"`
}

type DailyWeather struct {
	Date                     string  `json:"date"`
	WeatherCode              int     `json:"weatherCode"`
	Condition                string  `json:"condition"`
	TempMax                  float64 `json:"tempMax"`
	TempMin                  float64 `json:"tempMin"`
	PrecipitationProbability *int    `json:"precipitationProbability"`
}
