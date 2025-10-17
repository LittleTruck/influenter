package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/designcomb/influenter-backend/internal/database"
	"github.com/designcomb/influenter-backend/internal/utils"
	"github.com/designcomb/influenter-backend/internal/workers"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
)

func main() {
	// 1. 載入配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化結構化日誌
	logger := utils.InitLogger(cfg.Env, cfg.LogLevel)
	logger.Info().
		Str("env", cfg.Env).
		Msg("Config loaded successfully")

	// 3. 初始化加密工具
	if err := utils.InitCrypto(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize crypto")
	}
	logger.Info().Msg("Crypto initialized successfully")

	// 4. 連接資料庫
	db, err := database.New(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()
	logger.Info().
		Str("host", cfg.Database.Host).
		Str("database", cfg.Database.Database).
		Msg("Database connected successfully")

	// 5. 建立 Redis 連線選項
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	// 6. 建立 Asynq server
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6, // 高優先級任務
				"default":  3, // 一般任務
				"low":      1, // 低優先級任務
			},
			// 錯誤處理
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.Error().
					Err(err).
					Str("task_type", task.Type()).
					Msg("Task failed")
			}),
			// 日誌
			Logger: &asynqLogger{logger: &logger},
		},
	)

	// 7. 建立 Asynq client（用於 enqueue 任務）
	client := asynq.NewClient(redisOpt)
	defer client.Close()

	// 8. 建立 mux (任務路由)
	mux := asynq.NewServeMux()

	// 9. 註冊任務處理器
	mux.HandleFunc(workers.TypeEmailSync, func(ctx context.Context, t *asynq.Task) error {
		return workers.HandleEmailSyncTask(ctx, t, db.DB)
	})
	mux.HandleFunc(workers.TypeEmailSyncAll, func(ctx context.Context, t *asynq.Task) error {
		return workers.HandleEmailSyncAllTask(ctx, t, db.DB, client)
	})

	logger.Info().Msg("✅ Task handlers registered:")
	logger.Info().Msg("   - " + workers.TypeEmailSync)
	logger.Info().Msg("   - " + workers.TypeEmailSyncAll)

	// 10. 建立 Scheduler（定期任務）
	scheduler := asynq.NewScheduler(redisOpt, nil)

	// 註冊定期任務：每 5 分鐘同步所有使用者的郵件
	syncAllTask, err := workers.NewEmailSyncAllTask(100)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create sync all task")
	}

	if _, err := scheduler.Register("*/5 * * * *", syncAllTask); err != nil {
		logger.Fatal().Err(err).Msg("Failed to register scheduled task")
	}

	logger.Info().Msg("✅ Scheduled tasks registered:")
	logger.Info().Msg("   - Email sync all users (every 5 minutes)")

	// 11. 啟動 scheduler
	if err := scheduler.Start(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start scheduler")
	}
	defer scheduler.Shutdown()

	logger.Info().
		Str("redis_addr", cfg.Redis.Addr).
		Msg("🔄 Worker is starting...")

	// 12. 啟動 worker
	if err := srv.Start(mux); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start worker")
	}

	// 13. 等待中斷信號
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	// 14. 優雅關閉
	logger.Info().Msg("Shutting down worker...")
	srv.Shutdown()
	scheduler.Shutdown()
	time.Sleep(time.Second)
	logger.Info().Msg("Worker stopped")
}

// asynqLogger 實作 asynq.Logger 介面
type asynqLogger struct {
	logger *zerolog.Logger
}

func (l *asynqLogger) Debug(args ...interface{}) {
	l.logger.Debug().Msg(fmt.Sprint(args...))
}

func (l *asynqLogger) Info(args ...interface{}) {
	l.logger.Info().Msg(fmt.Sprint(args...))
}

func (l *asynqLogger) Warn(args ...interface{}) {
	l.logger.Warn().Msg(fmt.Sprint(args...))
}

func (l *asynqLogger) Error(args ...interface{}) {
	l.logger.Error().Msg(fmt.Sprint(args...))
}

func (l *asynqLogger) Fatal(args ...interface{}) {
	l.logger.Fatal().Msg(fmt.Sprint(args...))
}
