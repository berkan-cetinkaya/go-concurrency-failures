# Experiment 01 — Blocked Send (Inventory Reservation Hangs)

## Problem
In a high-throughput order management system, the HTTP handler hands off orders to a background shipment worker via an unbuffered channel to keep the response time low. However, the system starts "hanging" for users, and response times spike to the client-side timeout limits.

## Why it Happens
The background shipment worker crashes (e.g., failed connection to a shipping provider) and the goroutine exits silently. Since the handoff is done via an unbuffered channel (`shipments <- orderID`), the next HTTP request blocks indefinitely waiting for a receiver that no longer exists.

## How to Reproduce
1.  The `OrderService` starts a worker.
2.  The worker fails to connect to the shipping provider and exits.
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
- [ ] **TCP Connections**: (If the client doesn't timeout) HTTP connections stay open, potentially exhausting the server's file descriptors.

## How to Detect
- **Unit Testing**: Use `uber-go/goleak` to catch blocked goroutines at the end of your tests.
- **Metrics**: Monitor the number of active goroutines (`go_goroutines`). A steady upward slope without a plateau is a red flag.
- **Health Checks**: If your health check depends on a internal worker heartbeat, it should fail when the worker exits.

## How to Fix
Use a `select` statement with a `context` timeout to ensure the handler never waits indefinitely.

```go
select {
case s.shipments <- orderID:
    // Success
case <-ctx.Done():
    // Fail fast
    http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
}
```

## Related Concepts
- **Unbounded Handoff**: The danger of unbuffered channels in request paths.
- **Liveness vs. Readiness**: Why a process can be "alive" but "broken."
- **Backpressure**: The need to signal the client when the system can't keep up.
