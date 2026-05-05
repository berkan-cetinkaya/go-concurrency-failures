package main

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestOrderHandler_NoLeakInSolution(t *testing.T) {
	// This should PASS because the solution uses context timeouts
	defer goleak.VerifyNone(t)

	service := &OrderService{
		shipments: make(chan string),
	}
	// We actually call StartWorker(), but it will exit silently 
	// because connectToShippingProvider() fails.
	service.StartWorker()

	// Give the worker a moment to "crash"
	time.Sleep(50 * time.Millisecond)

	req := httptest.NewRequest("GET", "/order?id=SOLUTION-TEST", nil)
	w := httptest.NewRecorder()

	// 500ms timeout in our solution, so we wait slightly longer to be sure
	service.OrderHandler(w, req)

	fmt.Println("Test: Request finished. goleak will now verify that no goroutines are hanging.")
}
