package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/geoah/go-kube-api/internal/api"
	"github.com/geoah/go-kube-api/internal/rbac"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type config struct {
	BindAddress string `envconfig:"bind_address" default:"localhost:8080"`
}

func main() {
	// construct a logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println("error constructing logger, err:", err.Error())
		os.Exit(1)
	}
	// flushes buffer, if any
	defer logger.Sync()

	// parse configuration
	config := config{}
	if err := envconfig.Process("", &config); err != nil {
		logger.Fatal("error parsing config", zap.Error(err))
	}

	// construct an in-cluster config
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Fatal("error constructing in-cluster config", zap.Error(err))
	}

	// construct the clientset
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		logger.Fatal("error constructing clientset", zap.Error(err))
	}

	// construct RBAC enumerator
	rbacEnumerator, err := rbac.New(kubeClient.RbacV1())
	if err != nil {
		logger.Fatal("error constructing rbac enumerator", zap.Error(err))
	}

	// construct API
	api, err := api.New(rbacEnumerator)
	if err != nil {
		logger.Fatal("error constructing api", zap.Error(err))
	}

	// construct HTTP router
	router := gin.New()

	// add the ginzap middleware to make gin log through zap
	router.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(logger, true))

	// setup routes
	router.POST("/v1/rbac/enumerateBySubjectNames", api.RbacEnummerateByBindings)
	router.GET("/healthz", api.Health)

	// construct HTTP server
	srv := &http.Server{
		Addr:    config.BindAddress,
		Handler: router,
	}

	// start HTTP server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("error serving HTTP", zap.Error(err))
		}
	}()

	// wait for any signal that we should stop serving HTTP requests
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	logger.Info("shutting down HTTP server")

	// we allow 5 seconds for any remaining requests to be processed
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// shut down the server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("error while shutting down server", zap.Error(err))
	}

	// graceful shutdown completed
	logger.Info("server shut down")
}
