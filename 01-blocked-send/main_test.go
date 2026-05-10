package main

import (
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

func TestOrderHandler_TableDriven_LeakDetection(t *testing.T) {
	tests := []struct {
		name           string
		simulateFail   bool
		expectedStatus int
		leakReason     string
	}{
		{
			name:           "Failure_Scenario_Handler_Leaks",
			simulateFail:   true,
			expectedStatus: http.StatusOK,
			leakReason:     "Handler blocks forever waiting for a dead worker",
		},
		{
			name:           "Success_Scenario_Worker_Leaks",
			simulateFail:   false,
			expectedStatus: http.StatusOK,
			leakReason:     "Worker stays alive forever because there is no Stop/Context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The broken scenarios intentionally leave goroutines behind. Capture
			// the baseline so previous subtests do not hide the current leak.
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			fmt.Printf("%s[RUNNING]%s %s\n", colorBlue, colorReset, tt.name)

			service := &OrderService{
				shipments: make(chan string),
				provider:  &MockShippingProvider{ShouldFail: tt.simulateFail},
			}

			// Broken version starts worker without context
			service.StartWorker()
			time.Sleep(50 * time.Millisecond)

			req := httptest.NewRequest("GET", "/order?id=TEST-ORDER", nil)
			w := httptest.NewRecorder()

			if tt.simulateFail {
				// We KNOW this will block the handler goroutine
				go service.OrderHandler(w, req)
				time.Sleep(100 * time.Millisecond)
			} else {
				// Handler finishes normally
				service.OrderHandler(w, req)
			}
			// Use appropriate color for the log message
			color := colorGreen
			if tt.simulateFail {
				color = colorRed
			}
			fmt.Printf("%s[FINISHED]%s %s\n", color, colorReset, tt.name)
		})
	}
}
