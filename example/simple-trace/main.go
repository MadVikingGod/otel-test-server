package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

var (
	url   string
	count int
	help  bool
)

func init() {
	flag.StringVar(&url, "url", "localhost:30001", "The url the client will connect to")
	flag.StringVar(&url, "u", "localhost:30001", "The url the client will connect to (shorthand)")
	flag.IntVar(&count, "count", 1, "The Number of traces to send")
	flag.IntVar(&count, "c", 1, "The Number of traces to send (shorthand)")
	flag.BoolVar(&help, "help", false, "Display this help message")
	flag.BoolVar(&help, "h", false, "Display this help message")

}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
	}

	cp := grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  3 * time.Second,
			Multiplier: 2,
			Jitter:     0,
			MaxDelay:   120 * time.Second,
		},
		MinConnectTimeout: 10 * time.Second,
	}

	driver := otlpgrpc.NewDriver(
		otlpgrpc.WithEndpoint(url),
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithDialOption(grpc.WithConnectParams(cp)),
	)
	exp, err := otlp.NewExporter(context.Background(), driver)
	if err != nil {
		fmt.Println("Failed to create Exporter, ", err)
		os.Exit(1)
	}
	defer exp.Shutdown(context.Background())

	otel.SetTracerProvider(trace.NewTracerProvider(trace.WithSyncer(exp)))

	for i := 0; i < count; i++ {
		fmt.Println("Sending Span ", i)
		_, span := otel.Tracer("fake-tracer").Start(context.Background(), fmt.Sprintf("span-%d", i))

		span.End()

	}

}
