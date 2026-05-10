package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ShippingProvider defines the contract for external shipping services.
// This allows us to mock the provider in tests without changing the business logic.
type ShippingProvider interface {
	Connect() error
}

// RealShippingProvider is the production implementation that talks to the network.
type RealShippingProvider struct {
	URL string
}

func (p *RealShippingProvider) Connect() error {
	client := http.Client{Timeout: 1 * time.Second}
	_, err := client.Get(p.URL)
	if err != nil {
		return fmt.Errorf("shipping provider unreachable on %s", p.URL)
	}
	return nil
}

type OrderService struct {
	shipments chan string
	provider  ShippingProvider
}

func (s *OrderService) StartWorker(ctx context.Context) {
	go func() {
		fmt.Println("Worker: starting...")

		if err := s.provider.Connect(); err != nil {
			fmt.Printf("Worker: failed to connect to shipping provider: %v\n", err)
			return
		}

		fmt.Println("Worker: connected to shipping provider.")
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Worker: stopping due to context cancellation...")
				return
			case orderID, ok := <-s.shipments:
				if !ok {
					fmt.Println("Worker: channel closed, exiting...")
					return
				}
				fmt.Printf("Worker: creating shipment for order %s\n", orderID)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (s *OrderService) OrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("id")
	if orderID == "" {
		http.Error(w, "missing order id", http.StatusBadRequest)
		return
	}

	fmt.Printf("Handler: received order %s\n", orderID)

	ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
	defer cancel()

	fmt.Printf("Handler: sending order to shipment worker...\n")
	select {
	case s.shipments <- orderID:
		fmt.Fprintf(w, "Order %s accepted\n", orderID)
	case <-ctx.Done():
		fmt.Printf("Handler: failed to send order %s: %v\n", orderID, ctx.Err())
		http.Error(w, "Service Unavailable: background worker not responding", http.StatusServiceUnavailable)
	}
}

func main() {
	// Production uses the real provider
	provider := &RealShippingProvider{URL: "http://localhost:9999/health"}

	service := &OrderService{
		shipments: make(chan string),
		provider:  provider,
	}

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	service.StartWorker(workerCtx)

	mux := http.NewServeMux()
	mux.HandleFunc("/order", service.OrderHandler)
	server := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		fmt.Println("Server (FIXED) listening on http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	
	fmt.Println("\nShutdown signal received...")
	workerCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	fmt.Println("Server exiting gracefully.")
}
