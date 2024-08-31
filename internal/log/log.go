package log

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Setup(debug bool) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().
		Timestamp().
		Logger()

	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		logger = logger.With().Caller().Logger()
	}

	log.Logger = logger
}
