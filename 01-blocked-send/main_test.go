package main

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestOrderHandler_LeakDetection(t *testing.T) {
	// This will fail the test if any goroutines are leaked/blocked
	defer goleak.VerifyNone(t)

	service := &OrderService{
		shipments: make(chan string),
	}
	// We don't start the worker, simulating the crash

	req := httptest.NewRequest("GET", "/order?id=LEAK-TEST", nil)
	w := httptest.NewRecorder()

	// Run the handler in a goroutine because it will block
	go service.OrderHandler(w, req)

	// Give it a moment to block on the channel send
	time.Sleep(100 * time.Millisecond)

	// When the test ends, goleak will check for the blocked goroutine
	fmt.Println("Test: finishing and checking for leaks...")
}
