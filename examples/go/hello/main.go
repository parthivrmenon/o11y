package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		response := map[string]string{"msg": "Hello"}
		w.Header().Set("content-type", "application/json")
		json.NewEncoder(w).Encode(response)

	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

}

func initTracer() func() {
	// Create a stdout exporter to write traces to the console (for debugging)
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new tracer provider with the stdout exporter
	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String("hello-world-server"),
		)),
	)

	// Set the tracer provider globally
	otel.SetTracerProvider(tracerProvider)

	// Return a function to shutdown the tracer provider when done
	return func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	// init tracer
	shutdown := initTracer()
	defer shutdown()

	// Use otelhttp for instrumentation of the HTTP server
	wrappedHandler := otelhttp.NewHandler(http.HandlerFunc(HelloHandler), "/hello")

	// Register the /hello route with the OpenTelemetry-instrumented handler
	http.Handle("/hello", wrappedHandler)

	fmt.Println("Listening on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server failed to start:", err)
	}

}
