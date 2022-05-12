package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ds248a/nami/config"
	"github.com/ds248a/nami/log"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/patrickmn/go-cache"
)

var (
	cfg *config.Config
	srv *http.Server
	pdb *pgxpool.Pool
	rdb *redis.Ring
	lc  *cache.Cache
)

func init() {
	// Загрузка конфигурационного файла
	cf, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	cfg = cf

	// Log
	if err = log.NewLog(
		&log.Config{
			Debug:   cfg.Debug,
			Format:  cfg.Logger.Format,
			LogFile: cfg.Logger.LogFile,
		}); err != nil {
		log.Fatal(err)
	}
	callOnExit(log.Close)

	// Local Cache
	if err = newCache(cfg.Cache); err != nil {
		log.Fatal(err)
	}

	// Postgre
	if cfg.Postgre.Enable {
		if err = newPostgre(cfg.Postgre); err != nil {
			log.Fatal(err)
		}
	}

	// Redis DB
	if cfg.Redis.Enable {
		if err = newRedis(cfg.Redis); err != nil {
			log.Fatal(err)
		}
	}
}

// --------------------------------
//    Nami
// --------------------------------

// Роутер приложения.
func Router() *gin.Engine {
	gin.ForceConsoleColor()

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	// Список резрешенный прокси-серверов
	// []string{"172.16.0.10"} - пример списа
	// nil - если не используются
	r.SetTrustedProxies(nil) // no proxy
	return r
}

// Запуск HTTP сервера.
func StartHTTP(r *gin.Engine) {
	NewServer(r, cfg)
}

// --------------------------------
//    Close Connect
// --------------------------------

type hookFn func(context.Context, *sync.WaitGroup)

var onExit []hookFn

// Обработка прерываний сервера HTTP.
// Список значений: os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT
func Signal() os.Signal {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	return <-sigint
}

// Закрытие открытых соединений с ограничением по времени исполнения.
func Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Msg("test 1").Save()

	var wg sync.WaitGroup
	wg.Add(len(onExit))

	for _, h := range onExit {
		go func(h hookFn) {
			h(ctx, &wg)
		}(h)
	}

	log.Msg("test 3").Save()
	wg.Wait()
}

func callOnExit(h hookFn) {
	onExit = append(onExit, h)
}
