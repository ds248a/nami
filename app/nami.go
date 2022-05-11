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
	// загрузка конфигурационного файла
	cf, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	cfg = cf

	// Postgre
	//if cfg.Postgre.Enable {
	log.Msg("Postgre connect").Save()
	if err = newPostgre(cfg.Postgre); err != nil {
		log.Fatal(err)
	}
	//}

	// Redis DB
	if err = newRedis(cfg.Redis); err != nil {
		log.Fatal(err)
	}

	// Local Cache
	if err = newCache(cfg.Cache); err != nil {
		log.Fatal(err)
	}

	// log
	if err = log.NewLog(&log.Config{Debug: cfg.Debug, Format: cfg.Logger.Format, PDB: pdb}); err != nil {
		log.Fatal(err)
	}

	callOnExit(log.Close)
}

// --------------------------------
//    Nami
// --------------------------------

// роутер приложения
func Router() *gin.Engine {
	gin.ForceConsoleColor()

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	// r.SetTrustedProxies([]string{"192.168.1.2"})
	r.SetTrustedProxies(nil) // no proxy
	return r
}

// запуск HTTP сервера
func StartHTTP(r *gin.Engine) {
	NewServer(r, cfg)
}

// --------------------------------
//    Close Connect
// --------------------------------

type hookFn func(context.Context, *sync.WaitGroup)

var onExit []hookFn

// обработка прерываний сервера HTTP
// список значений: os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT
func Signal() os.Signal {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	return <-sigint
}

// закрытие открытых соединений с ограничением по времени исполнения
func Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// log.go - "test 2"
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
