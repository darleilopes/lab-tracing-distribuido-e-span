package cep

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestFetchCityNameByCEPSuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	expectedCity := "São Paulo"
	httpmock.RegisterResponder("GET", "http://viacep.com.br/ws/05025000/json/",
		httpmock.NewStringResponder(200, `{"localidade": "`+expectedCity+`"}`))

	cityName, err := FetchCityNameByCEP("05025000", context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cityName != expectedCity {
		t.Fatalf("expected cityName to be %s, got %s", expectedCity, cityName)
	}
}

func TestFetchCityNameByCEPNotFound(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://viacep.com.br/ws/00000000/json/",
		httpmock.NewStringResponder(404, ""))

	_, err := FetchCityNameByCEP("00000000", context.Background())
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}
