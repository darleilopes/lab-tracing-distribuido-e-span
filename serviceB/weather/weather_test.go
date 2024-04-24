package weather

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestFetchTemperatureByCityNameSuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	encodedCity := url.QueryEscape("São Paulo")
	expectedTemp := 22.0
	httpmock.RegisterResponder("GET", fmt.Sprintf("http://api.weatherapi.com/v1/current.json?q=%s&key=%s", encodedCity, WeatherAPIKey),
		httpmock.NewStringResponder(200, `{"current": {"temp_c": `+fmt.Sprintf("%f", expectedTemp)+`}}`))

	tempC, err := FetchTemperatureByCityName("São Paulo", context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tempC != expectedTemp {
		t.Fatalf("expected tempC to be %f, got %f", expectedTemp, tempC)
	}
}

func TestFetchTemperatureByCityNameNotFound(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", fmt.Sprintf("http://api.weatherapi.com/v1/current.json?q=InvalidCity&key=%s", WeatherAPIKey),
		httpmock.NewStringResponder(404, ""))

	_, err := FetchTemperatureByCityName("InvalidCity", context.Background())
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}
