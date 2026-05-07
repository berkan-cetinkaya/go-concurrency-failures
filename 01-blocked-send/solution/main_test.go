package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/goleak"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorBlue   = "\033[34m"
	colorYellow = "\033[33m"
)

// MockShippingProvider allows us to simulate failures or successes
type MockShippingProvider struct {
	ShouldFail bool
}

func (m *MockShippingProvider) Connect() error {
	if m.ShouldFail {
		return fmt.Errorf("connection refused (mocked)")
	}
	return nil
}

func TestOrderHandler_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		simulateFail   bool
		expectedStatus int
	}{
		{
			name:           "Failure_Scenario_Worker_Dead",
			simulateFail:   true,
			expectedStatus: http.StatusServiceUnavailable, // 503
		},
		{
			name:           "Success_Scenario_Worker_Alive",
			simulateFail:   false,
			expectedStatus: http.StatusOK, // 200
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Granular leak check
			defer goleak.VerifyNone(t)

			fmt.Printf("%s[RUNNING]%s %s\n", colorBlue, colorReset, tt.name)

			service := &OrderService{
				shipments: make(chan string),
				provider:  &MockShippingProvider{ShouldFail: tt.simulateFail},
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			service.StartWorker(ctx)

			// Give the worker a moment to start or fail
			time.Sleep(50 * time.Millisecond)

			req := httptest.NewRequest("GET", "/order?id=TEST-ORDER", nil)
			w := httptest.NewRecorder()

			service.OrderHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("%sFAILED%s: %s: expected status %d, got %d", colorRed, colorReset, tt.name, tt.expectedStatus, w.Code)
			}
			
			color := colorGreen
			if tt.simulateFail {
				color = colorYellow
			}
			fmt.Printf("%s[FINISHED]%s %s with status %d\n", color, colorReset, tt.name, w.Code)
		})
	}
}
