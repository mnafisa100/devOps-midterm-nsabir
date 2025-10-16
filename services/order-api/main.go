
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Order struct {
	ID         int       `json:"id"`
	CustomerID int       `json:"customer_id"`
	ProductID  int       `json:"product_id"`
	Quantity   int       `json:"quantity"`
	Total      float64   `json:"total"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Count   int         `json:"count,omitempty"`
}

var (
	orders      = make(map[int]*Order)
	ordersMutex = &sync.RWMutex{}
	nextID      = 1
	startTime   = time.Now()
)

func main() {
	initOrders()
	
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/ready", readyHandler)
	http.HandleFunc("/api/orders", ordersHandler)
	http.HandleFunc("/api/orders/", orderHandler)
	http.HandleFunc("/metrics", metricsHandler)
	http.HandleFunc("/", rootHandler)
	
	port := getEnv("PORT", "8080")
	log.Printf("✅ Order API starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func initOrders() {
	orders[1] = &Order{
		ID: 1, CustomerID: 101, ProductID: 1,
		Quantity: 2, Total: 1999.98, Status: "completed",
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}
	orders[2] = &Order{
		ID: 2, CustomerID: 102, ProductID: 3,
		Quantity: 1, Total: 79.99, Status: "pending",
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}
	nextID = 3
	log.Printf("Initialized %d sample orders", len(orders))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"service":   "order-api",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(startTime).Seconds(),
	})
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ready",
		"service": "order-api",
	})
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		getOrders(w)
	case "POST":
		createOrder(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getOrders(w http.ResponseWriter) {
	ordersMutex.RLock()
	defer ordersMutex.RUnlock()
	
	list := make([]*Order, 0, len(orders))
	for _, o := range orders {
		list = append(list, o)
	}
	
	log.Printf("Fetching all orders - Total: %d", len(list))
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Count:   len(list),
		Data:    list,
	})
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	if order.CustomerID == 0 || order.ProductID == 0 || order.Quantity == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	
	ordersMutex.Lock()
	order.ID = nextID
	nextID++
	order.CreatedAt = time.Now()
	order.Status = "pending"
	orders[order.ID] = &order
	ordersMutex.Unlock()
	
	log.Printf("Order created: %d", order.ID)
	
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Success: true, Data: order})
}

func orderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.URL.Path[len("/api/orders/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}
	
	switch r.Method {
	case "GET":
		getOrder(w, id)
	case "PUT":
		updateOrder(w, r, id)
	case "DELETE":
		deleteOrder(w, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getOrder(w http.ResponseWriter, id int) {
	ordersMutex.RLock()
	order, exists := orders[id]
	ordersMutex.RUnlock()
	
	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Data: order})
}

func updateOrder(w http.ResponseWriter, r *http.Request, id int) {
	ordersMutex.Lock()
	defer ordersMutex.Unlock()
	
	order, exists := orders[id]
	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	
	var updates Order
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	if updates.Status != "" {
		order.Status = updates.Status
	}
	if updates.Quantity > 0 {
		order.Quantity = updates.Quantity
	}
	
	log.Printf("Order updated: %d", id)
	json.NewEncoder(w).Encode(Response{Success: true, Data: order})
}

func deleteOrder(w http.ResponseWriter, id int) {
	ordersMutex.Lock()
	defer ordersMutex.Unlock()
	
	if _, exists := orders[id]; !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	
	delete(orders, id)
	log.Printf("Order deleted: %d", id)
	
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    map[string]string{"message": fmt.Sprintf("Order %d deleted", id)},
	})
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	ordersMutex.RLock()
	count := len(orders)
	ordersMutex.RUnlock()
	
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "# HELP orders_total Total orders\n")
	fmt.Fprintf(w, "# TYPE orders_total gauge\n")
	fmt.Fprintf(w, "orders_total %d\n", count)
	fmt.Fprintf(w, "\n# HELP app_uptime_seconds Application uptime\n")
	fmt.Fprintf(w, "# TYPE app_uptime_seconds gauge\n")
	fmt.Fprintf(w, "app_uptime_seconds %.2f\n", time.Since(startTime).Seconds())
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "Order API",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"health":  "/health",
			"ready":   "/ready",
			"orders":  "/api/orders",
			"metrics": "/metrics",
		},
	})
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
