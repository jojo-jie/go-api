package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/sync/singleflight"
	"log"
	"net/http"
	"time"
)

var (
	cost  float64
	group *singleflight.Group
)

//When to Use (and When Not to Use)
//何时使用（以及何时不使用）
//Use:   使用：
//For read operations (queries to APIs or databases).
//对于读取操作（对 API 或数据库的查询）。
//When the resource is idempotent (does not change the system state).
//当资源幂等（不会改变系统状态）。
//Do not use:   不要使用：
//For write operations (creation, update, deletion).
//对于写操作（创建、更新、删除）。
//When the result may vary between calls.
//当结果可能在调用之间变化时。

func init() {
	group = new(singleflight.Group)
}

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

	price, err, _ := group.Do(productID, func() (interface{}, error) {
		return fetchProductPrice(productID)
	})
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
	response := map[string]interface{}{
		"total_cost": fmt.Sprintf("%.2f", cost),
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func clearCosts(w http.ResponseWriter, r *http.Request) {
	cost = 0
}
