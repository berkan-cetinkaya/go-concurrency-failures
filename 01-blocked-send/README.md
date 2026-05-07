# Experiment 01 — Blocked Send (Shipment Worker Handoff Hangs)

## Problem
In a high-throughput order management system, the HTTP handler hands off orders to a background shipment worker via an unbuffered channel to keep the response time low. However, the system starts "hanging" for users, and response times spike to the client-side timeout limits.

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
- [x] **Goroutines**: Every incoming request creates a new goroutine that never exits.
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

### 2. Graceful Shutdown & Lifecycle Management
Use `context.Context` to manage the worker's lifecycle. Ensure that when the application shuts down, all background workers are notified and closed properly.

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
