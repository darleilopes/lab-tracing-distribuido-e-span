package logHandler

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"log"
)

func HandleLogError(message string, err error, ctx context.Context) {
	tr := otel.Tracer("service-a-error-tracer")
	ctx, errorHandlingSpan := tr.Start(ctx, "Error handling")
	defer errorHandlingSpan.End()

	errorHandlingSpan.SetAttributes(attribute.KeyValue{
		Key:   "error_message",
		Value: attribute.StringValue(message),
	})

	if err == nil {
		log.Println(message)
		return
	} else {
		log.Println(message, err.Error())
		errorHandlingSpan.SetAttributes(attribute.KeyValue{
			Key:   "error",
			Value: attribute.StringValue(err.Error()),
		})
	}
}
