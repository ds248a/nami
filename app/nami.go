package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ds248a/air/config"
	lg "github.com/ds248a/air/log"
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
	if err = lg.NewLog(cfg.Loger, pdb); err != nil {
		log.Fatal(err)
	}

	callOnExit(lg.LogClose)
}

// --------------------------------
//    Nami
// --------------------------------

// роутер приложения
func Router() *gin.Engine {
	return gin.Default() // debugPrintWARNINGDefault() + engine.Use(Logger(), Recovery())
	// return gim.New()
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
	// lg.LogMsg("test 1").Save()

	var wg sync.WaitGroup
	wg.Add(len(onExit))

	for _, h := range onExit {
		go func(h hookFn) {
			h(ctx, &wg)
		}(h)
	}

	// lg.LogMsg("test 3").Save()

	wg.Wait()
}

func callOnExit(h hookFn) {
	onExit = append(onExit, h)
}
