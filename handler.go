package httpecho

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type HttpEchoHandler struct {
	logger    zerolog.Logger
	jwtHeader string
	counter   uint64
	mutex     sync.Mutex
}

func (h *HttpEchoHandler) writeErr(w http.ResponseWriter, err error, msg string) {
	h.logger.Error().Err(err).Msg(msg)
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"error":   true,
		"message": msg,
		"details": err.Error(),
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(response); err != nil {
		h.logger.Error().Err(err).Msg("failed to write error JSON")
	}
}

func (h *HttpEchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var counter uint64 = 0
	h.mutex.Lock()
	h.counter += 1
	counter = h.counter
	h.mutex.Unlock()

	h.logger.Info().Uint64("counter", counter).Msgf("serving %s %s to %s", r.Method, r.RequestURI, r.RemoteAddr)

	response := HttpEchoResponse{
		StatusCode: http.StatusOK,
		Hostname:   r.Host,
		Headers:    r.Header,
		Path:       r.URL.Path,
		Method:     r.Method,
		Query:      r.URL.Query(),
		Body:       nil,
		Json:       nil,
		Jwt:        nil,
	}

	if r.ContentLength > 0 {
		if r.Header.Get("Content-Type") == "application/json" {
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&response.Json); err != nil {
				h.writeErr(w, err, "failed decoding json body")
				return
			}
		} else {
			buf, err := io.ReadAll(r.Body)
			if err != nil {
				h.writeErr(w, err, "failed reading body")
				return
			}
			body := string(buf)
			response.Body = &body
		}
	}

	if h.jwtHeader != "" {
		jwtRaw := r.Header.Get(h.jwtHeader)
		if jwtRaw != "" {
			jwtRawParts := strings.Split(jwtRaw, " ")
			i := min(1, len(jwtRawParts))
			jwt, err := ParseJwtString(jwtRawParts[i])
			if err != nil {
				h.writeErr(w, err, "failed to parse jwt")
				return
			}
			response.Jwt = jwt
		}
	}

	if v := r.Header.Get("X-Set-Response-Status-Code"); v != "" {
		if c, err := strconv.Atoi(v); err != nil || c < 1 {
			h.logger.Warn().Err(err).Msg("invalid status code requested")
		} else {
			h.logger.Info().Uint64("counter", counter).Msgf("status code forced to %d", c)
			response.StatusCode = c
		}
	}

	if v := r.Header.Get("X-Set-Response-Delay-Ms"); v != "" {
		if c, err := strconv.Atoi(v); err != nil || c < 1 {
			h.logger.Warn().Err(err).Msg("invalid delay ms requested")
		} else {
			d := time.Millisecond * time.Duration(c)
			h.logger.Info().Uint64("counter", counter).Msgf("sleeping for to %v", d)
			time.Sleep(d)
		}
	}

	w.WriteHeader(response.StatusCode)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&response); err != nil {
		h.writeErr(w, err, "failed to marshal response as JSON")
		return
	}

	h.logger.Info().Uint64("counter", counter).Any("response", response).Msgf("served %s %s to %s", r.Method, r.RequestURI, r.RemoteAddr)
}

func NewHttpEchoHandler(logger zerolog.Logger, jwtHeader string) *HttpEchoHandler {
	return &HttpEchoHandler{
		logger:    logger,
		jwtHeader: jwtHeader,
		counter:   0,
		mutex:     sync.Mutex{},
	}
}
