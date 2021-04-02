package retry

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	pbctrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	count       int
	minDuration time.Duration
	maxDuration time.Duration

	testString string
	errorCode  codes.Code

	msgs    chan string
	started chan struct{}

	pbctrace.UnimplementedTraceServiceServer
}

func New(name string, count int, min, max time.Duration, errCode codes.Code) *Handler {
	h := &Handler{
		count:       count,
		minDuration: min,
		maxDuration: max,

		testString: name,
		errorCode:  errCode,

		msgs:    make(chan string),
		started: make(chan struct{}, 1),
	}
	h.started <- struct{}{}
	return h
}

func (h *Handler) Export(ctx context.Context, req *pbctrace.ExportTraceServiceRequest) (*pbctrace.ExportTraceServiceResponse, error) {
	id, err := getID(req)
	if err != nil {
		return nil, status.Error(codes.Internal, "request did not have ID")
	}
	select {
	case <-h.started:
		go h.startTest(ctx, id)
		return nil, status.Error(h.errorCode, "test failed successfully")
	case <-ctx.Done():
		return nil, status.Error(codes.Internal, "context cancled")
	default:
	}
	h.msgs <- id

	return nil, status.Error(h.errorCode, "test failed successfully")
}

func getID(req *pbctrace.ExportTraceServiceRequest) (string, error) {
	if len(req.ResourceSpans) < 1 ||
		len(req.ResourceSpans[0].InstrumentationLibrarySpans) < 1 ||
		len(req.ResourceSpans[0].InstrumentationLibrarySpans[0].Spans) < 1 {
		return "", fmt.Errorf("requst did not have an ID")
	}
	return string(req.ResourceSpans[0].InstrumentationLibrarySpans[0].Spans[0].TraceId), nil
}

func (h *Handler) startTest(ctx context.Context, id string) {
	defer func() {
		//Drain msgs
		for {
			select {
			case <-h.msgs:
			default:
				h.started <- struct{}{}
				return
			}
		}
	}()
	startTime := time.Now()
	log := log.WithField("testName", h.testString)
	log.Info("starting test")
	failed := false
	timeout := time.NewTimer(h.maxDuration)
	count := 0
	for {
		select {
		case <-timeout.C:
			timeout.Stop()
			goto done
		case reqID := <-h.msgs:
			log := log
			if reqID != id {
				// If IDs don't match it's not a retry
				continue
			}
			if time.Now().Before(startTime.Add(h.minDuration)) {
				log = log.WithField("earlyRetry", true)
				failed = true
			}
			count++
			log.WithField("count", count).Debug("retry")
		}
	}
done:
	if failed {
		fmt.Printf("Failed: %s - Retried too soon\n", h.testString)
		return
	}
	if h.count != count {
		fmt.Printf("Failed: %s - Too many/few Retries. Expected %d, got %d\n", h.testString, h.count, count)
		return
	}
	fmt.Printf("Success: %s\n", h.testString)

}
