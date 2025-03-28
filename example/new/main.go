package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
	"time"
)

var (
	cost      float64
	group     *singleflight.Group
	templates *template.Template
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
	templatePath := filepath.Join("templates", "index.html")
	var err error
	templates, err = template.ParseFiles(templatePath)
	if err != nil {
		panic(err)
	}
}

type SseHandler struct {
	clients map[chan string]struct{}
}

func NewSseHandler() *SseHandler {
	return &SseHandler{
		clients: make(map[chan string]struct{}),
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templates.Execute(w, nil)
	})
	mux.HandleFunc("/products/{id}/price", getProductPriceHandler)
	mux.HandleFunc("/costs", getCost)
	mux.HandleFunc("/clear-costs", clearCosts)
	sseHandler := NewSseHandler()
	mux.HandleFunc("/sse", sseHandler.Serve)
	g := new(errgroup.Group)
	g.Go(func() error {
		s := &http.Server{
			Addr:           ":8080",
			Handler:        mux,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   5 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		defer s.Close()
		log.Println("API running without single flight on port :8080...")
		return s.ListenAndServe()
	})

	g.Go(func() error {
		return sseHandler.SimulateEvents()
	})

	if err := g.Wait(); err != nil {
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

func (h *SseHandler) Serve(w http.ResponseWriter, r *http.Request) {
	// 设置 SSE 必需的 HTTP 头
	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Connection", "keep-alive")
	clientChan := make(chan string)
	h.clients[clientChan] = struct{}{}

	defer func() {
		delete(h.clients, clientChan)
		close(clientChan)
	}()

	for {
		select {
		case msg := <-clientChan:
			//使用SSE格式，data: 后接消息内容，结尾需两个换行符（\n\n），以标志消息结束。客户端（如浏览器的EventSource）会解析此格式，触发消息事件。
			fmt.Fprintf(w, "data: %s\n\n", msg)
			//调用 Flush() 确保数据即时发送，而非等待缓冲区满或请求结束。这对实时性要求高的场景（如SSE）至关重要。
			w.(http.Flusher).Flush()
		case <-r.Context().Done():

			return
		}
	}
}

func (h *SseHandler) SimulateEvents() error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		message := fmt.Sprintf("Server time: %s Total Cost: %s", time.Now().Format(time.RFC3339), fmt.Sprintf("%.2f", cost))
		for clientChan := range h.clients {
			select {
			case clientChan <- message:
			default:
				// 跳过阻塞的 channel
			}
		}
	}
	return nil
}
