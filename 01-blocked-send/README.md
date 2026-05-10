# Experiment 01 — Blocked Send (Shipment Worker Handoff Hangs)

## Problem
In a high-throughput order management system, the HTTP handler synchronously hands off orders to a background shipment worker via an unbuffered channel. If the worker is available this looks simple, but if the worker is gone the request path blocks directly. The system starts "hanging" for users, and response times spike to the client-side timeout limits.

## Why it Happens
The background shipment worker fails to start properly (e.g., failed connection to a shipping provider) and the goroutine exits silently. Since the handoff is done via an unbuffered channel (`shipments <- orderID`), the next HTTP request blocks indefinitely waiting for a receiver that no longer exists.

## How to Reproduce
1.  The `OrderService` starts a worker.
2.  The worker fails its initial connection (natural network failure or simulated via `MockShippingProvider`) and exits.
3.  The `OrderHandler` receives a request and attempts to send to the channel.
4.  The request hangs forever.

```bash
cd 01-blocked-send
make reproduce
```

## Symptoms in Production
- **Response Time**: Spikes to the maximum client-side or Load Balancer timeout.
- **pprof**: A growing number of goroutines stuck in `chan send` state.
- **CPU Usage**: Low. Since the goroutines are blocked and not "spinning," the CPU doesn't spike. This is why **Autoscaling (HPA)** based on CPU will fail to trigger.
- **Memory**: Gradual increase as each blocked request keeps its stack, context, and associated objects in memory.
- **Goroutine Dump**: Look for `goroutine [chan send]: ...` pointing to the line where you send to the shipment channel.

## What actually Leaked?
- [x] **Goroutines**: `net/http` runs each incoming request in its own goroutine, and each blocked send pins that request goroutine forever.
- [x] **Memory**: Request contexts and order objects are pinned in memory.
- [x] **Worker Slots**: The background worker is gone, but the system doesn't know it.

## How to Detect
- **Unit Testing**: Use `uber-go/goleak` to catch blocked goroutines at the end of your tests.
- **Metrics**: Monitor the number of active goroutines (`go_goroutines`). A steady upward slope without a plateau is a red flag.
- **Health Checks**: If your health check depends on an internal worker heartbeat, it should fail when the worker exits.

## How to Fix

### 1. Fail Fast with Context Timeout
Never allow a front-facing handler to wait indefinitely on an internal channel. Use `select` with a context timeout.

```go
select {
case s.shipments <- orderID:
    // Success
case <-ctx.Done():
    // Fail fast
    http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
}
```

This protects the request path from hanging forever, but it does not make the worker healthy. If the worker failed during startup, the service may still return `503` for every request until the process is restarted or the worker is recovered.

### 2. Worker Lifecycle and Readiness
Use `context.Context` to manage the worker's lifecycle, and make worker startup failure visible to the rest of the process. A production service should either fail startup, mark readiness as failed, or supervise/restart the worker when the required dependency is unavailable.

```go
func (s *OrderService) StartWorker(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return // Stop worker
            case job := <-s.shipments:
                // Process job
            }
        }
    }()
}
```

## Related Concepts
- **Unbounded Handoff**: The danger of unbuffered channels in request paths.
- **Graceful Shutdown**: Listening for SIGINT/SIGTERM to stop background work.
- **Liveness vs. Readiness**: Why a process can be "alive" but "broken."
