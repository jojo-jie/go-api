package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/hibiken/asynq"
	"io"
	"math"
	"net/http"
	"os"
	"redenvelop/redpacket"
	"strconv"
	"time"
)

const localRedisAddr = "127.0.0.1:6379"

var rdb *redis.Client

var c *asynq.Client

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr: localRedisAddr,
	})

	cif := asynq.RedisClientOpt{Addr: localRedisAddr, DB: 13}
	srv := asynq.NewServer(
		cif, asynq.Config{LogLevel: asynq.FatalLevel},
	)
	c = asynq.NewClient(cif)
	err := srv.Start(asynq.HandlerFunc(Job))
	if err != nil {
		panic(err)
	}
}

const redList = 8

const redKey = "red:list"

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/red_list", createRedList).Methods(http.MethodGet)
	r.HandleFunc("/red", getRed).Methods(http.MethodGet)
	r.HandleFunc("/rob_red", robRed).Methods(http.MethodGet)
	s := &http.Server{
		Addr:           ":4498",
		Handler:        r,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.ListenAndServe()
}

func Job(ctx context.Context, task *asynq.Task) error {
	fmt.Println(task.Type())
	return nil
}

func createRedList(writer http.ResponseWriter, request *http.Request) {
	num := request.URL.Query().Get("num")
	totalPrice := request.URL.Query().Get("total_price")
	redNum := redList
	totalP := float64(13.56)
	if num != "" {
		redNum, _ = strconv.Atoi(num)
	}
	if totalPrice != "" {
		totalP, _ = strconv.ParseFloat(totalPrice, 64)
	}
	l, err := redpacket.NewRedPacketPool(totalP, redNum)
	if err != nil {
		writer.Write([]byte(err.Error()))
		return
	}

	rdb.Del(request.Context(), redKey)
	list := make(map[int]redpacket.RedPacket, 8)
	for k, packet := range l {
		p, _ := json.Marshal(packet)
		s := string(p)
		rdb.RPush(context.Background(), redKey, s).Err()
		list[k] = *packet
	}

	result := Result{
		Code: 0,
		Data: list,
	}
	res, _ := json.Marshal(result)
	writer.Write(res)
}

type Data struct {
	Total float64                     `json:"total"`
	List  map[int]redpacket.RedPacket `json:"list"`
}

type Result struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

// 获取红包
func getRed(writer http.ResponseWriter, request *http.Request) {
	l := rdb.LRange(context.Background(), redKey, 0, -1).Val()
	sum := float64(0)

	list := make(map[int]redpacket.RedPacket, len(l))
	for k, s := range l {
		v := &redpacket.RedPacket{}
		json.Unmarshal([]byte(s), v)
		list[k] = *v
		sum += math.Ceil(v.Amount * 100.0)
	}
	result := Result{
		Code: 0,
		Data: Data{
			Total: sum / 100,
			List:  list,
		},
	}

	res, _ := json.Marshal(result)
	writer.Write(res)
}

const userHashKey = "user:prevent:replay"
const userAmountKey = "user:amount:log"

// 抢红包
func robRed(writer http.ResponseWriter, request *http.Request) {
	f, _ := os.Open("./lua/red.lua")
	redScript, _ := io.ReadAll(f)
	lua := redis.NewScript(string(redScript))
	userId := request.URL.Query().Get("user_id")
	lua.Load(request.Context(), rdb)
	result, _ := lua.EvalSha(request.Context(), rdb, []string{userHashKey, redKey, userAmountKey, userId}).Result()
	res := []byte(result.(string))
	c.Enqueue(asynq.NewTask("save_rob_red_result", res, asynq.MaxRetry(3), asynq.Timeout(3*time.Second), asynq.ProcessIn(10*time.Second)))
	writer.Write(res)
}
