# otel-test-server
A server to test opentelemetry implementation against.  This server is intended to validate behavior of OTLP clients.  


# Current Tests

All tests assume that you have an OTLP client that is configured to send 1 or more requests to the collector.  If the tests need a configuration different from the spec it should be detailed under the test.

## Retries
The [Open Telemetry Protocol Spec](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/otlp.md#otlpgrpc-response) allows cleints to retry requests when it receives a number of different error codes.  These retries must also backoff with an exponential decay.

The server will start a ["trace collector"](https://github.com/open-telemetry/opentelemetry-proto/tree/main/opentelemetry/proto/collector/trace/v1) on ports 30001-30016 that only respons with the corrisponding errorcode. 

### Test conditions
INITIAL_BACKOFF = 1 second
MIN_CONNECT_TIMEOUT = 20 seconds
MULTIPLIER = 1.6
MAX_BACKOFF = 120 seconds
JITTER = 0.2

The client should send 1 message, and it should retry the same message 5 times.  After 30 second if the correct number of retries happened the server should report a success.