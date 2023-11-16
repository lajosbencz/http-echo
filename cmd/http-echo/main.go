package main

import (
	"context"
	"crypto/tls"
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

	log0 "log"

	httpecho "github.com/lajosbencz/http-echo"
	selfsignedtlsgo "github.com/lajosbencz/selfsigned-tls-go"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	envLogJson     = "LOG_JSON"
	envListenHost  = "LISTEN_HOST"
	envListenHttp  = "LISTEN_HTTP"
	envListenHttps = "LISTEN_HTTPS"
	envCorsEnabled = "CORS_ENABLED"
	envJwtEnabled  = "JWT_ENABLED"
	envJwtHeader   = "JWT_HEADER"
	envLogLevel    = "LOG_LEVEL"
)

var (
	envEnabled  = false
	logJson     = false
	listenHost  = "0.0.0.0"
	listenHttp  = 8080
	listenHttps = 8443
	corsEnabled = false
	jwtEnabled  = false
	jwtHeader   = "Authorization"
	logLevel    = int(zerolog.InfoLevel)
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

func init() {
	flag.BoolVar(&envEnabled, "env", envEnabled, "Overwrite options from ENV")
	flag.BoolVar(&logJson, "log-json", logJson, "Set log format to JSON")
	flag.StringVar(&listenHost, "host", listenHost, "Host to listen on")
	flag.IntVar(&listenHttp, "http", listenHttp, "HTTP port to listen on")
	flag.IntVar(&listenHttps, "https", listenHttps, "HTTPS port to listen on, 0 turns it off")
	flag.BoolVar(&corsEnabled, "cors", corsEnabled, "Allow CORS")
	flag.BoolVar(&jwtEnabled, "jwt", jwtEnabled, "Enable parsing of JWT")
	flag.StringVar(&jwtHeader, "jwt-header", jwtHeader, "JWT header name")
	flag.IntVar(&logLevel, "log-level", logLevel, "Logging level of Zerolog")
	flag.Parse()

	if envEnabled {
		if httpecho.GetEnvBool(envLogJson) {
			logJson = true
		}
		logLevel = httpecho.GetEnvInt(envLogLevel, logLevel)
	}

	if !logJson {
		log.Logger = log.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.TimeFormat = "2006-01-02 15:04:05"
		}))
	}
	zerolog.SetGlobalLevel(zerolog.Level(logLevel))

	if envEnabled {
		if envHost := os.Getenv(envListenHost); envHost != "" {
			listenHost = envHost
		}

		listenHttp = httpecho.GetEnvInt(envListenHttp, listenHttp)

		listenHttps = httpecho.GetEnvInt(envListenHttps, listenHttps)

		if httpecho.GetEnvBool(envCorsEnabled) {
			corsEnabled = true
		}

		if httpecho.GetEnvBool(envJwtEnabled) {
			jwtEnabled = true
		}

		if jwtHeaderEnv := os.Getenv(envJwtHeader); jwtHeaderEnv != "" {
			jwtHeader = jwtHeaderEnv
		}
	}
}

func main() {

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGHUP)
	defer close(shutdown)

	defer log.Debug().Msg("shut down gracefully")

	log.Info().Bool("enabled", envEnabled).Msg("ENV")
	log.Info().Bool("enabled", corsEnabled).Msg("CORS")
	log.Info().Bool("enabled", jwtEnabled).Msg("JWT")

	jwtFinalHeader := ""
	if jwtEnabled {
		jwtFinalHeader = jwtHeader
	}

	echoHandler := httpecho.NewHttpEchoHandler(log.Logger, jwtFinalHeader)

	muxHandler := http.NewServeMux()
	muxHandler.Handle("/", echoHandler)
	muxHandler.HandleFunc("/favicon.ico", handleFavicon)

	serverErr := make(chan error, 1)

	var finalHandler http.Handler
	if corsEnabled {
		finalHandler = cors.AllowAll().Handler(muxHandler)
	} else {
		finalHandler = muxHandler
	}

	listenAddr := fmt.Sprintf("%s:%d", listenHost, listenHttp)

	log0.Default().SetFlags(log0.Lshortfile)
	httpLogWriter := &zerologWriter{
		logger: log.Logger,
	}
	httpLogger := log0.New(httpLogWriter, "", 0)

	serverHttp := &http.Server{
		Addr:     listenAddr,
		Handler:  finalHandler,
		ErrorLog: httpLogger,
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := serverHttp.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("http: %s", err)
		} else {
			log.Debug().Msg("server (http) closed")
		}
	}()

	var serverHttps *http.Server
	if listenHttps > 0 {
		// @todo: check if crt files were provided
		listenAddrHttps := fmt.Sprintf("%s:%d", listenHost, listenHttps)
		if crt, err := selfsignedtlsgo.DefaultSelfsignedTls(); err != nil {
			log.Fatal().Err(err).Msg("failed to generate self-signed TLS")
		} else {
			serverHttps = &http.Server{
				Addr:    listenAddrHttps,
				Handler: finalHandler,
				TLSConfig: &tls.Config{
					NextProtos:         []string{"http/1.1"},
					Certificates:       []tls.Certificate{crt},
					InsecureSkipVerify: true,
				},
				ErrorLog: httpLogger,
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := serverHttps.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
					serverErr <- fmt.Errorf("https: %s", err)
				} else {
					log.Debug().Msg("server (https) closed")
				}
			}()
		}
		log.Info().Str("address", listenAddrHttps).Msgf("server (https) listening")
		defer log.Debug().Msg("server (https) stopped")
	}

	log.Info().Str("address", listenAddr).Msgf("server (http) listening")
	defer log.Debug().Msg("server (http) stopped")

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

	log.Debug().Msg("shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := serverHttp.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("server (http) graceful shutdown error")
		}
	}()

	if serverHttps != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := serverHttps.Shutdown(ctx); err != nil {
				log.Error().Err(err).Msg("server (https) graceful shutdown error")
			}
		}()
	}

	wg.Wait()
}
