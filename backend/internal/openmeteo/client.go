package openmeteo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"weatherwatch/internal/forecast"
	"weatherwatch/internal/watchlist"
)

const (
	defaultGeocodingURL = "https://geocoding-api.open-meteo.com/v1/search"
	defaultForecastURL  = "https://api.open-meteo.com/v1/forecast"
)

type Client struct {
	HTTP         *http.Client
	GeocodingURL string
	ForecastURL  string
}

func New() *Client {
	return &Client{
		HTTP:         &http.Client{Timeout: 10 * time.Second},
		GeocodingURL: defaultGeocodingURL,
		ForecastURL:  defaultForecastURL,
	}
}

type rawPlace struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Admin1    string  `json:"admin1"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (c *Client) SearchPlaces(ctx context.Context, query string, limit int) ([]watchlist.Place, error) {
	params := url.Values{
		"name":     {query},
		"count":    {strconv.Itoa(limit)},
		"language": {"en"},
		"format":   {"json"},
	}
	// The API omits "results" entirely when nothing matches.
	var body struct {
		Results []rawPlace `json:"results"`
	}
	if err := c.getJSON(ctx, c.GeocodingURL, params, &body); err != nil {
		return nil, fmt.Errorf("search places: %w", err)
	}

	places := make([]watchlist.Place, len(body.Results))
	for i, p := range body.Results {
		places[i] = watchlist.Place(p)
	}
	return places, nil
}

type rawForecast struct {
	Timezone string `json:"timezone"`
	Current  struct {
		Time                string  `json:"time"`
		Temperature         float64 `json:"temperature_2m"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		Humidity            int     `json:"relative_humidity_2m"`
		WindSpeed           float64 `json:"wind_speed_10m"`
		WeatherCode         int     `json:"weather_code"`
		IsDay               int     `json:"is_day"`
	} `json:"current"`
	Daily struct {
		Time                     []string  `json:"time"`
		WeatherCode              []int     `json:"weather_code"`
		TempMax                  []float64 `json:"temperature_2m_max"`
		TempMin                  []float64 `json:"temperature_2m_min"`
		PrecipitationProbability []*int    `json:"precipitation_probability_max"`
	} `json:"daily"`
}

func (c *Client) Forecast(ctx context.Context, lat, lon float64) (forecast.Forecast, error) {
	params := url.Values{
		"latitude":      {strconv.FormatFloat(lat, 'f', 4, 64)},
		"longitude":     {strconv.FormatFloat(lon, 'f', 4, 64)},
		"current":       {"temperature_2m,apparent_temperature,relative_humidity_2m,wind_speed_10m,weather_code,is_day"},
		"daily":         {"weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max"},
		"timezone":      {"auto"},
		"forecast_days": {"7"},
	}
	var raw rawForecast
	if err := c.getJSON(ctx, c.ForecastURL, params, &raw); err != nil {
		return forecast.Forecast{}, fmt.Errorf("forecast: %w", err)
	}
	return raw.toDomain()
}

func (r rawForecast) toDomain() (forecast.Forecast, error) {
	d := r.Daily
	n := len(d.Time)
	if len(d.WeatherCode) != n || len(d.TempMax) != n || len(d.TempMin) != n || len(d.PrecipitationProbability) != n {
		return forecast.Forecast{}, fmt.Errorf("forecast: daily series have mismatched lengths")
	}

	days := make([]forecast.DailyWeather, n)
	for i := range n {
		days[i] = forecast.DailyWeather{
			Date:                     d.Time[i],
			WeatherCode:              d.WeatherCode[i],
			Condition:                describeWeatherCode(d.WeatherCode[i]),
			TempMax:                  d.TempMax[i],
			TempMin:                  d.TempMin[i],
			PrecipitationProbability: d.PrecipitationProbability[i],
		}
	}

	cur := r.Current
	return forecast.Forecast{
		Timezone: r.Timezone,
		Current: forecast.CurrentWeather{
			Time:                cur.Time,
			Temperature:         cur.Temperature,
			ApparentTemperature: cur.ApparentTemperature,
			Humidity:            cur.Humidity,
			WindSpeed:           cur.WindSpeed,
			WeatherCode:         cur.WeatherCode,
			Condition:           describeWeatherCode(cur.WeatherCode),
			IsDay:               cur.IsDay == 1,
		},
		Daily: days,
	}, nil
}

func (c *Client) getJSON(ctx context.Context, base string, params url.Values, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"?"+params.Encode(), nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("upstream returned %d: %s", resp.StatusCode, snippet)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<16)).Decode(dst); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
