package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var cost float64

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/products/{id}/price", getProductPriceHandler)
	mux.HandleFunc("/costs", getCost)
	mux.HandleFunc("/clear-costs", clearCosts)
	s := &http.Server{
		Addr:           ":8080",
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("API running without singleflight on port :8080...")
	err := s.ListenAndServe()
	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}

func fetchProductPrice(productID string) (float64, error) {
	log.Printf("[COST: $0.01] Calling external API for product: %s\n", productID)
	time.Sleep(2 * time.Second) // Simulates latency
	cost += 0.01
	return 99.99, nil
}

func getProductPriceHandler(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("id")

	price, err := fetchProductPrice(productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"product_id": productID,
		"price":      price,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getCost(w http.ResponseWriter, r *http.Request) {

}

func clearCosts(w http.ResponseWriter, r *http.Request) {

}
