package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
)

func main() {
	// Redis 連線設定
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	// 建立 Asynq server
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	// 建立 mux (任務路由)
	mux := asynq.NewServeMux()

	// TODO: 註冊任務處理器
	// mux.HandleFunc("email:sync", handleEmailSync)
	// mux.HandleFunc("email:analyze", handleEmailAnalysis)

	log.Println("🔄 Worker is starting...")
	log.Printf("📡 Connected to Redis at %s", redisAddr)

	// 啟動 worker
	if err := srv.Start(mux); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	// 等待中斷信號
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	// 優雅關閉
	log.Println("Shutting down worker...")
	srv.Shutdown()
	time.Sleep(time.Second)
	log.Println("Worker stopped")
}

