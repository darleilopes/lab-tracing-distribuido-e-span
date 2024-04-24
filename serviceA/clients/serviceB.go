package serviceB

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

type ServiceBAPIResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func CallServiceB(cep string, ctx context.Context) (*ServiceBAPIResponse, error) {
	tr := otel.Tracer("service-a-cep-handler-tracer")
	ctx, serviceBSpan := tr.Start(ctx, "callServiceB")
	defer serviceBSpan.End()

	serviceBURL := "http://service_b:8082/temperature"

	req, err := http.NewRequestWithContext(ctx, "GET", serviceBURL+"?cep="+cep, nil)
	if err != nil {
		return nil, fmt.Errorf("Erro ao criar a requisição para o Serviço B: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Erro ao fazer a requisição para o Serviço B: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("Serviço B retornou status: %d - Cannot find zipcode data", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		return nil, fmt.Errorf("Serviço B retornou status: %d - Invalid zipcode", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Serviço B retornou status: %d", resp.StatusCode)
	}

	var data ServiceBAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("Erro ao ler a resposta do Serviço B: %v", err)
	}

	return &data, nil
}
