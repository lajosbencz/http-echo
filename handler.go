package httpecho

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

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
		Hostname: r.Host,
		Headers:  r.Header,
		Uri:      r.RequestURI,
		Method:   r.Method,
		Query:    r.URL.Query(),
		Body:     nil,
		Json:     nil,
		Jwt:      nil,
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

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&response); err != nil {
		h.writeErr(w, err, "failed to marshal response as JSON")
		return
	}

	h.logger.Info().Int64("counter", int64(counter)).Any("response", response).Msgf("served %s %s to %s", r.Method, r.RequestURI, r.RemoteAddr)
}

func NewHttpEchoHandler(logger zerolog.Logger, jwtHeader string) *HttpEchoHandler {
	return &HttpEchoHandler{
		logger:    logger,
		jwtHeader: jwtHeader,
		counter:   0,
		mutex:     sync.Mutex{},
	}
}
