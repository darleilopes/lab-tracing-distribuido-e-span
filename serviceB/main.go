package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/darleilopes/lab-tracing-distribuido-e-span/serviceB/cep"
	"github.com/darleilopes/lab-tracing-distribuido-e-span/serviceB/configs"
	"github.com/darleilopes/lab-tracing-distribuido-e-span/serviceB/weather"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.7.0"
)

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
			semconv.ServiceNameKey.String("Service-B"),
		)),
	)

	otel.SetTracerProvider(tp)
}

type TemperatureResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
	City  string  `json:"city"`
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
	http.Handle("/temperature", otelhttp.NewHandler(http.HandlerFunc(handleTemperatureRequest), "TemperatureEndpoint"))

	log.Println("Server starting on port 8082...")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleTemperatureRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tr := otel.Tracer("service-b-temperature-tracer")

	ctx, temperatureSpan := tr.Start(ctx, "handleTemperatureRequest")
	defer temperatureSpan.End()

	if r.Method != http.MethodGet {
		message := "Method is not supported"
		logHandler.HandleLogError(message, nil, ctx)
		http.Error(w, message, http.StatusNotFound)
		return
	}

	query := r.URL.Query()
	zipcode := query.Get("cep")

	temperatureSpan.SetAttributes(attribute.KeyValue{
		Key:   "zipcode",
		Value: attribute.StringValue(zipcode),
	})

	if len(zipcode) != 8 {
		message := "Invalid zipcode"
		logHandler.HandleLogError(message, nil, ctx)
		http.Error(w, message, http.StatusUnprocessableEntity)
		return
	}

	cityName, err := cep.FetchCityNameByCEP(zipcode, ctx)
	if err != nil {
		message := "Can not find zipcode"
		logHandler.HandleLogError(message, err, ctx)
		http.Error(w, message, http.StatusNotFound)
		return
	}

	tempC, err := weather.FetchTemperatureByCityName(cityName, ctx)
	if err != nil {
		message := "Error fetching temperature"
		logHandler.HandleLogError(message, err, ctx)
		http.Error(w, message, http.StatusInternalServerError)
		return
	}

	response := TemperatureResponse{
		TempC: tempC,
		TempF: tempC*1.8 + 32,
		TempK: tempC + 273.15,
		City:  cityName,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
