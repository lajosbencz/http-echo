package main

import (
	"github.com/rs/zerolog"
)

type zerologWriter struct {
	logger zerolog.Logger
}

func (w *zerologWriter) Write(bytes []byte) (n int, err error) {
	w.logger.Warn().Msg(string(bytes))
	return len(bytes), nil
}
