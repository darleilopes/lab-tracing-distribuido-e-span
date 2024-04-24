package cep

import (
	"context"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"net/http"
)

type ViaCEPResponse struct {
	Localidade string `json:"localidade"`
}

func FetchCityNameByCEP(cep string, ctx context.Context) (string, error) {
	tr := otel.Tracer("service-b-temperature-tracer")
	ctx, fetchCityNameByCEPSpan := tr.Start(ctx, "FetchCityNameByCEP")
	defer fetchCityNameByCEPSpan.End()

	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("Erro ao criar a requisição para o ViaCep: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("CEP not found: %v", resp.StatusCode)
	}

	var data ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if data.Localidade == "" {
		return "", fmt.Errorf("CEP not found: Sem localidade")
	}

	return data.Localidade, nil
}
