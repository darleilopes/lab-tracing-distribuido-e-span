package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/darleilopes/lab-tracing-distribuido-e-span/serviceB/weather"
	"github.com/jarcoal/httpmock"
)

type APIResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func TestTemperatureEndpoint(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var cityName = "São Paulo"
	var encodedCity = url.QueryEscape(cityName)

	httpmock.RegisterResponder("GET", "http://viacep.com.br/ws/05025000/json/",
		httpmock.NewStringResponder(200, `{"cep": "05025000", "localidade": "`+cityName+`"}`))

	httpmock.RegisterResponder("GET", fmt.Sprintf("http://api.weatherapi.com/v1/current.json?q=%s&key=%s", encodedCity, weather.WeatherAPIKey),
		httpmock.NewStringResponder(200, `{"current": {"temp_c": 22.0}}`))

	req, err := http.NewRequest("GET", "/temperature?cep=05025000", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleTemperatureRequest)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := APIResponse{TempC: 22.0, TempF: 71.6, TempK: 295.15}

	var data APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&data); err != nil {
		t.Errorf("handler returned unexpected body: Cannot convert response to JSON")
	}
	if data != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", data, expected)
	}
}

func TestTemperatureEndpointWithInvalidCEP(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	req, err := http.NewRequest("GET", "/temperature?cep=invalid", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleTemperatureRequest)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnprocessableEntity)
	}
}

func TestTemperatureEndpointCEPNotFound(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://viacep.com.br/ws/00000000/json/",
		httpmock.NewStringResponder(404, ""))

	req, err := http.NewRequest("GET", "/temperature?cep=00000000", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleTemperatureRequest)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestTemperatureEndpointWeatherAPIFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var cityName = "São Paulo"
	var encodedCity = url.QueryEscape(cityName)

	httpmock.RegisterResponder("GET", "http://viacep.com.br/ws/01001000/json/",
		httpmock.NewStringResponder(200, `{"cep": "01001000", "localidade": "`+cityName+`"}`))

	httpmock.RegisterResponder("GET", fmt.Sprintf("http://api.weatherapi.com/v1/current.json?q=%s&key=%s", encodedCity, weather.WeatherAPIKey),
		httpmock.NewStringResponder(500, ""))

	req, err := http.NewRequest("GET", "/temperature?cep=01001000", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleTemperatureRequest)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}
