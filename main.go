package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"testserver/retry"

	pbctrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func main() {
	fmt.Println("Starting Test Server")
	for code, tt := range RetryServers {
		name := fmt.Sprintf("RetryTest - %s (%d)", code.String(), int(code))
		handler := retry.New(name, tt.count, tt.minTime, tt.maxTime, code)
		port := 30000 + int(code)
		fmt.Printf("Start - :%d - %s\n", port, name)
		newOtelServer(handler, port)
	}
	for true {
	}
}

var RetryServers = map[codes.Code]struct {
	count   int
	minTime time.Duration
	maxTime time.Duration
}{
	codes.Canceled:           {5, 5 * time.Second, 30 * time.Second},
	codes.Unknown:            {0, 0, 30 * time.Second},
	codes.InvalidArgument:    {0, 0, 30 * time.Second},
	codes.DeadlineExceeded:   {5, 5 * time.Second, 30 * time.Second},
	codes.NotFound:           {0, 0, 30 * time.Second},
	codes.AlreadyExists:      {0, 0, 30 * time.Second},
	codes.PermissionDenied:   {5, 5 * time.Second, 30 * time.Second},
	codes.ResourceExhausted:  {5, 5 * time.Second, 30 * time.Second},
	codes.FailedPrecondition: {0, 0, 30 * time.Second},
	codes.Aborted:            {5, 5 * time.Second, 30 * time.Second},
	codes.OutOfRange:         {5, 5 * time.Second, 30 * time.Second},
	codes.Unimplemented:      {0, 0, 30 * time.Second},
	codes.Internal:           {0, 0, 30 * time.Second},
	codes.Unavailable:        {5, 5 * time.Second, 30 * time.Second},
	codes.DataLoss:           {5, 5 * time.Second, 30 * time.Second},
	codes.Unauthenticated:    {5, 5 * time.Second, 30 * time.Second},
}

type ExportFunc func(context.Context, *pbctrace.ExportTraceServiceRequest) (*pbctrace.ExportTraceServiceResponse, error)

func newOtelServer(handler pbctrace.TraceServiceServer, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	serv := grpc.NewServer()
	pbctrace.RegisterTraceServiceServer(serv, handler)
	go serv.Serve(lis)
	return nil
}
