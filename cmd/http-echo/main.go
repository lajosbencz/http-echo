package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	httpecho "github.com/lajosbencz/http-echo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	listenHost = "0.0.0.0"
	listenPort = 8080
	jwtHeader  = ""
)

func main() {
	log.Logger = log.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = "2006-01-02 15:04:05"
	}))

	defer log.Info().Msg("stopped.")

	flag.StringVar(&listenHost, "host", listenHost, "Host to listen on")
	flag.IntVar(&listenPort, "port", listenPort, "Port to listen on")
	flag.StringVar(&jwtHeader, "jwt-header", jwtHeader, "JWT header name")
	flag.Parse()

	if envPort := os.Getenv("LISTEN_PORT"); envPort != "" {
		p, err := strconv.Atoi(envPort)
		if err != nil {
			log.Logger.Fatal().Err(err).Msg("failed to parse port")
		}
		listenPort = p
	}

	if jwtHeaderEnv := os.Getenv("JWT_HEADER"); jwtHeaderEnv != "" {
		jwtHeader = jwtHeaderEnv
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGHUP)

	listenAddr := fmt.Sprintf("%s:%d", listenHost, listenPort)

	handler := httpecho.NewHttpEchoHandler(log.Logger, jwtHeader)

	server := &http.Server{
		Addr:    listenAddr,
		Handler: handler,
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("server listen error")
		} else {
			log.Info().Msg("server closed")
		}
	}()

	log.Info().Msgf("server listening on %s", listenAddr)

	<-shutdown

	log.Info().Msg("graceful stop...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("error while server shutdown")
	}

	wg.Wait()
}
