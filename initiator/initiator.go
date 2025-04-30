package initiator

import (
	"context"
	"digital-wallet/initiator/domain"
	"digital-wallet/initiator/foundation"
	"digital-wallet/initiator/platform"
	"digital-wallet/internal/constant/model/persistencedb"
	modu "digital-wallet/internal/module/catch"

	"digital-wallet/internal/handler/middleware"
	"digital-wallet/platform/logger"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func Initiate() {
	sampleLogger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf(`{"level":"fatal","msg":"failed to initialize sample logger: %v"}`, err)
		os.Exit(1)
	}

	sampleLogger.Info("initializing config")
	configName := "config"
	if name := os.Getenv("CONFIG_NAME"); name != "" {
		configName = name
		sampleLogger.Info(fmt.Sprintf("config name is set to %s", configName))
	} else {
		sampleLogger.Info("using default config name 'config'")
	}
	foundation.InitConfig(configName, "config", sampleLogger)
	sampleLogger.Info("config initialized")
	log := logger.New(platform.InitLogger())
	log.Info(context.Background(), "logger initialized")
	log.Info(context.Background(), "initializing database")
	pgxConn := foundation.InitDB(viper.GetString("database.url"), log)
	log.Info(context.Background(), "database initialized")

	log.Info(context.Background(), "initializing cache")
	redis := foundation.InitCache(viper.GetString("redis.url"), log)
	log.Info(context.Background(), "cache initialized")

	if viper.GetBool("migration.active") {
		log.Info(context.Background(), "initializing migration")
		m := foundation.InitiateMigration(viper.GetString("migration.path"),
			viper.GetString("database.url"), log)
		foundation.UpMigration(m, log)
		log.Info(context.Background(), "migration initialized")
	}

	log.Info(context.Background(), "initializing state")
	state := foundation.InitState(log)
	log.Info(context.Background(), "state initialized")

	cacheLayer := foundation.InitCacheLayer(foundation.CacheOptions{Redis: redis}, log)
	log.Info(context.Background(), "initializing persistance")
	persistenceLayer := domain.InitPersistence(persistencedb.New(pgxConn, log),
		log, *cacheLayer)
	log.Info(context.Background(), "persistance initialized")
	log.Info(context.Background(), "initializing platform layer")
	platformLayer := platform.InitHTTPClient(state.HTTPConfig, log)
	log.Info(context.Background(), "platform layer initialized")

	log.Info(context.Background(), "initializing module")
	module := domain.InitModule(persistenceLayer, log, platformLayer)
	log.Info(context.Background(), "module initialized")

	log.Info(context.Background(), "initializing handler")
	handler := domain.InitHandler(module, log, viper.GetDuration("server.timeout"))
	log.Info(context.Background(), "handler initialized")

	log.Info(context.Background(), "initializing server")
	server := gin.New()
	gin.SetMode(gin.ReleaseMode)
	server.Use(middleware.GinLogger(log.Named("gin")))
	server.Use(middleware.RecoveryWithZap(log.Named("gin.recovery"), true))
	server.Use(middleware.ErrorHandler())
	server.Use(domain.InitCORS())
	log.Info(context.Background(), "server initialized")
	v1 := server.Group("/v1")

	domain.InitRouter(v1, handler, log, module, state, *redis)
	log.Info(context.Background(), "router initialized")
	srv := &http.Server{
		Addr:              viper.GetString("server.host") + ":" + viper.GetString("server.port"),
		ReadHeaderTimeout: viper.GetDuration("read_header_timeout"),
		Handler:           server,
	}
	worker := modu.Init(log.Named("catch"), cacheLayer.Redis, nil, viper.GetViper().GetDuration(""), nil)
	go func() {
		if err := worker.Start(context.Background()); err != nil {
			log.Info(context.Background(), "Worker failed", zap.Error(err))
		}
	}()
	defer worker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)

	go func() {
		log.Info(context.Background(), "server started",
			zap.String("host", viper.GetString("server.host")),
			zap.Int("port", viper.GetInt("server.port")))
		log.Info(context.Background(), fmt.Sprintf("server stopped with error %v", srv.ListenAndServe()))
	}()
	sig := <-quit
	log.Info(context.Background(), fmt.Sprintf("server shutting down with signal %v", sig))
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("server.timeout"))
	defer cancel()

	log.Info(ctx, "shutting down server")
	err = srv.Shutdown(ctx)
	if err != nil {
		log.Fatal(context.Background(), fmt.Sprintf("error while shutting down server: %v", err))
	}
	log.Info(context.Background(), "server shutdown complete")

}
