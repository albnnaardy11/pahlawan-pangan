package middleware

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var tracer = otel.Tracer("http-middleware")

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// AnomalyDetectionMiddleware implements Tail-based Sampling logic
// It decides whether to sample (log/trace) a request based on its outcome (latency, error, status).
func AnomalyDetectionMiddleware(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 1. Start Span (initially unsampled/low priority if possible, but OTel sampling happens at start usually)
		// For tail-based, we often need a custom sampler or collector.
		// Here we simulate by controlling LOG verbosity and Span Attributes dynamically.
		ctx, span := tracer.Start(r.Context(), r.URL.Path)
		defer span.End()

		// Wrap ResponseWriter
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Buffer request body specific size for anomaly detection if needed (omitted for stream safety)
		// ...

		// 2. Process Request
		next.ServeHTTP(rw, r.WithContext(ctx))

		duration := time.Since(start)

		// 3. Dynamic Sampling Logic (The "Smart" Part)
		isAnomaly := false

		// Condition A: Slow Request (> 500ms)
		if duration > 500*time.Millisecond {
			isAnomaly = true
			span.SetAttributes(attribute.Bool("anomaly.slow", true))
		}

		// Condition B: Server Error (5xx)
		if rw.statusCode >= 500 {
			isAnomaly = true
			span.SetAttributes(attribute.Bool("anomaly.error", true))
			span.SetStatus(codes.Error, http.StatusText(rw.statusCode))
		}

		// Condition C: Payload Size (Content-Length) > 1MB
		if r.ContentLength > 1024*1024 {
			isAnomaly = true
			span.SetAttributes(attribute.Bool("anomaly.large_payload", true))
		}

		// 4. Structured Non-Blocking Logging (Zero Allocation where possible)
		fields := []zapcore.Field{
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rw.statusCode),
			zap.Duration("latency", duration),
			zap.String("trace_id", span.SpanContext().TraceID().String()),
		}

		if isAnomaly {
			// Force HIGH PRIORITY logging for anomalies
			// In a real tail-sampling system, this would signal the collector to KEEP the trace.
			span.SetAttributes(attribute.Bool("sampling.priority", true))
			logger.Error("🚨 Anomaly Detected", fields...)
		} else {
			// Sampling for Success (200 OK) -> Only log 0.1%
			// We use a simple probabilistic check here for the LOGS.
			// Traces are usually handled by the SDK sampler, but we can hint via events.
			if time.Now().UnixNano()%1000 == 0 { // ~0.1% chance
				logger.Info("✅ Request Processed (Sampled)", fields...)
			} else {
				// Mute logs for 99.9% of healthy requests to save I/O
			}
		}
	})
}
