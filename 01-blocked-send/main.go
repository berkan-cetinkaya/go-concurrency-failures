package main

import (
	"fmt"
	"log"
	"net/http"
)

type OrderService struct {
	shipments chan string
}

func (s *OrderService) StartWorker() {
	go func() {
		fmt.Println("Worker: starting...")

		if err := connectToShippingProvider(); err != nil {
			fmt.Printf("Worker: failed to connect to shipping provider: %v\n", err)
			return
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
	fmt.Println("Handler: sending order to shipment worker...")

	// This blocks forever because the worker exited during startup.
	// The channel is still open, but there is no receiver anymore.
	s.shipments <- orderID

	fmt.Fprintf(w, "Order %s accepted\n", orderID)
}

func connectToShippingProvider() error {
	return fmt.Errorf("connection refused")
}

func main() {
	service := &OrderService{
		shipments: make(chan string), // unbuffered channel
	}

	service.StartWorker()

	http.HandleFunc("/order", service.OrderHandler)

	fmt.Println("Server listening on http://localhost:8080")
	fmt.Println("Try: curl 'http://localhost:8080/order?id=ORD-001'")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
