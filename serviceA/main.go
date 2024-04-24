package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	serviceB "github.com/darleilopes/lab-tracing-distribuido-e-span/serviceA/clients"
	logHandler "github.com/darleilopes/lab-tracing-distribuido-e-span/serviceA/configs"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.7.0"
)

type CepRequest struct {
	CEP string `json:"cep"`
}

func initTracer() {
	exporter, err := zipkin.New(
		"http://zipkin:9411/api/v2/spans",
	)
	if err != nil {
		log.Fatalf("Failed to create Zipkin exporter: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("Service-A"),
		)),
	)

	otel.SetTracerProvider(tp)
}

func initPropagator() {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
}

func main() {
	initPropagator()
	initTracer()
	http.Handle("/cep", otelhttp.NewHandler(http.HandlerFunc(cepHandler), "CepEndpoint"))

	log.Println("Server starting on port 8081...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func cepHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tr := otel.Tracer("service-a-cep-handler-tracer")

	ctx, cepSpan := tr.Start(ctx, "cepHandler")
	defer cepSpan.End()

	if r.Method != http.MethodPost {
		message := "Método não suportado"
		logHandler.HandleLogError(message, nil, ctx)
		http.Error(w, message, http.StatusMethodNotAllowed)
		return
	}

	var req CepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		message := "Erro ao decodificar o corpo da requisição"
		logHandler.HandleLogError(message, err, ctx)
		http.Error(w, message, http.StatusBadRequest)
		return
	}

	cep := getCepFromInput(req.CEP)

	cepSpan.SetAttributes(attribute.KeyValue{
		Key:   "zipcode",
		Value: attribute.StringValue(cep),
	})

	if len(cep) != 8 {
		message := "Invalid zipcode"
		logHandler.HandleLogError(message, nil, ctx)
		http.Error(w, message, http.StatusUnprocessableEntity)
		return
	}

	response, err := serviceB.CallServiceB(req.CEP, ctx)
	if err != nil {
		message := fmt.Sprintf("Erro ao chamar o serviço B: %v", err)
		logHandler.HandleLogError(message, err, ctx)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getCepFromInput(input string) string {
	regex := regexp.MustCompile(`/\D/g`)
	rex := regex.ReplaceAllString(input, "")
	return strings.ReplaceAll(rex, " ", "")
}
