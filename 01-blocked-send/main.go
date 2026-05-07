package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type ShippingProvider interface {
	Connect() error
}

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

func (s *OrderService) StartWorker() {
	go func() {
		fmt.Println("Worker: starting...")

		if err := s.provider.Connect(); err != nil {
			fmt.Printf("Worker: failed to connect to shipping provider: %v\n", err)
			return // Worker exits silently!
		}

		fmt.Println("Worker: connected to shipping provider.")
		for orderID := range s.shipments {
			fmt.Printf("Worker: creating shipment for order %s\n", orderID)
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
	fmt.Printf("Handler: sending order to shipment worker...\n")

	// DEADLOCK: This blocks forever because the worker exited during startup.
	s.shipments <- orderID

	fmt.Fprintf(w, "Order %s accepted\n", orderID)
}

func main() {
	service := &OrderService{
		shipments: make(chan string),
		provider:  &RealShippingProvider{URL: "http://localhost:9999/health"},
	}

	service.StartWorker()

	http.HandleFunc("/order", service.OrderHandler)

	fmt.Println("Server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
