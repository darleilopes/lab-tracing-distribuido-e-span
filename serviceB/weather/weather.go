package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

const WeatherAPIKey = "d30f2cf2a4d949db83132323242404"

type WeatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

func FetchTemperatureByCityName(cityName string, ctx context.Context) (float64, error) {
	tr := otel.Tracer("service-b-temperature-tracer")
	ctx, fetchTemperatureByCityNameSpan := tr.Start(ctx, "FetchTemperatureByCityName")
	defer fetchTemperatureByCityNameSpan.End()

	encodedCityName := url.QueryEscape(cityName)
	requestURL := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?q=%s&key=%s", encodedCityName, WeatherAPIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return 0, fmt.Errorf("Erro ao criar a requisição para o WeatherAPI: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("weather information for city not found")
	}

	var data WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	return data.Current.TempC, nil
}
