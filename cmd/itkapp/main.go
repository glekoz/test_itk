package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/glekoz/test_itk/config"
	"github.com/glekoz/test_itk/internal/cache"
	"github.com/glekoz/test_itk/internal/repository"
	"github.com/glekoz/test_itk/internal/service"
	"github.com/glekoz/test_itk/internal/web/v1"
)

func main() {
	cfg := config.MustLoad()

	infoLog := log.New(os.Stdout, "[INFO]\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "[ERROR]\t", log.Ldate|log.Ltime|log.Lshortfile)

	repo, err := repository.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("can not connect to DB")
	}
	cache, err := cache.New(cfg.CacheTTL)
	if err != nil {
		log.Fatal("can not create cache")
	}
	s := service.New(repo, cache, infoLog, errorLog)
	server := web.New(s, cfg.Host, infoLog, errorLog)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		ErrorLog:     errorLog,
		Handler:      server.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("listening :%s", cfg.Port)

	if err := srv.ListenAndServe(); err != nil {
		errorLog.Fatal(err)
	}
}
