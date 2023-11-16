package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	httpecho "github.com/lajosbencz/http-echo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	envLogJson    = "LOG_JSON"
	envListenHost = "LISTEN_HOST"
	envListenPort = "LISTEN_PORT"
	envJwtEnabled = "JWT_ENABLED"
	envJwtHeader  = "JWT_HEADER"
)

var (
	logJson    = false
	listenHost = "0.0.0.0"
	listenPort = 8080
	jwtEnabled = false
	jwtHeader  = "Authorization"
)

//go:embed favicon.ico
var faviconFile embed.FS

func handleFavicon(w http.ResponseWriter, r *http.Request) {
	faviconData, err := faviconFile.ReadFile("favicon.ico")
	if err != nil {
		http.Error(w, "Favicon not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "image/x-icon")
	_, _ = w.Write(faviconData)
}

func main() {

	flag.BoolVar(&logJson, "log-json", logJson, "Set log format to JSON")
	flag.StringVar(&listenHost, "host", listenHost, "Host to listen on")
	flag.IntVar(&listenPort, "port", listenPort, "Port to listen on")
	flag.BoolVar(&jwtEnabled, "jwt", jwtEnabled, "Enable parsing of JWT")
	flag.StringVar(&jwtHeader, "jwt-header", jwtHeader, "JWT header name")
	flag.Parse()

	if httpecho.GetEnvBool(envLogJson) {
		logJson = true
	}

	if !logJson {
		log.Logger = log.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.TimeFormat = "2006-01-02 15:04:05"
		}))
	}

	if envHost := os.Getenv(envListenHost); envHost != "" {
		listenHost = envHost
	}

	listenPort = httpecho.GetEnvInt(envListenPort, listenPort)

	if httpecho.GetEnvBool(envJwtEnabled) {
		jwtEnabled = true
	}

	if jwtHeaderEnv := os.Getenv(envJwtHeader); jwtHeaderEnv != "" {
		jwtHeader = jwtHeaderEnv
	}

	jwtFinalHeader := ""
	if jwtEnabled {
		jwtFinalHeader = jwtHeader
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGHUP)

	listenAddr := fmt.Sprintf("%s:%d", listenHost, listenPort)

	handler := httpecho.NewHttpEchoHandler(log.Logger, jwtFinalHeader)

	mux := http.NewServeMux()
	mux.Handle("/", handler)
	mux.HandleFunc("/favicon.ico", handleFavicon)

	serverErr := make(chan error, 1)

	server := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		} else {
			log.Info().Msg("server closed")
		}
	}()

	log.Info().Msgf("server listening on %s", listenAddr)
	defer log.Info().Msg("server stopped")

loop:
	for {
		select {
		case <-shutdown:
			break loop
		case err := <-serverErr:
			log.Error().Err(err).Msg("server error")
			defer os.Exit(1)
			break loop
		}
	}

	log.Info().Msg("server graceful shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server graceful shutdown error")
		defer os.Exit(1)
	}

	wg.Wait()
}
