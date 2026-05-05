# Go Concurrency Failures

Production-grade failure patterns in Go: blocked channels, leaked goroutines, forgotten timeouts, broken cancellation, and backpressure collapse.

> **Concurrency failure is a waiting topology problem.**

This repository is a collection of real-world Go concurrency failures—ranging from blocked channels and leaked goroutines to circular lock dependencies and context ignorance. Each experiment is designed to show you not just the code, but the **production symptoms** you'll see in your monitoring tools.

## The Experiment Template
Each experiment in this lab follows a standardized high-value format:
- **Problem**: The business scenario (e.g., "Inventory reservation hangs").
- **Why it happens**: The technical root cause.
- **Symptoms in Production**: What you'll see in **pprof**, goroutine dumps, and why **Autoscaling** might fail to save you.
- **What actually leaked?**: A checklist of exhausted resources (Goroutines, TCP connections, DB slots).
- **How to detect**: Using tools like `goleak`, pprof, or custom metrics. (Note: We are keeping an eye on the upcoming `GOEXPERIMENT=goroutineleakprofile` feature in Go 1.26 for future inclusion).
- **How to fix**: The "Production-Grade" solution.
- **Related concepts**: Links to further reading.

## Experiments

1. ✅ [01-blocked-send](./01-blocked-send)
2. 🚧 [02-missing-receiver](./02-missing-receiver)
3. 🚧 [03-worker-startup-failure](./03-worker-startup-failure)
4. 🚧 [04-forgotten-timeout](./04-forgotten-timeout)
5. 🚧 [05-fan-in-close-panic](./05-fan-in-close-panic)
6. 🚧 [06-waitgroup-leak](./06-waitgroup-leak)
7. 🚧 [07-context-cancellation-ignored](./07-context-cancellation-ignored)
8. 🚧 [08-backpressure-collapse](./08-backpressure-collapse)
9. 🚧 [09-circular-lock-order](./09-circular-lock-order)
10. 🚧 [10-unbounded-worker-pool](./10-unbounded-worker-pool)

---

## Philosophy
This isn't just a tutorial. It's a practical reference for engineers who want to recognize concurrency failure patterns before they become production incidents.
