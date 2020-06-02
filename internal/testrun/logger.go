package testrun

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (t *TestRun) InitLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	logLevel := zerolog.InfoLevel

	if _, exist := os.LookupEnv("TESTRUN_DEBUG"); exist {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)
}

func logInfo(message string, content interface{}) {
	log.Info().Interface("content", content).Msg(message)
}

func logDebug(message string, content interface{}) {
	log.Debug().Interface("content", content).Msg(message)
}

//func logTest(message string, content interface{}) {
//	log.Debug().Interface("content", content).Msg(message)
//	log.Info().Interface("content", content).Msg(message)
//}

func logFatalError(err error, message string, content interface{}) error {
	log.Fatal().Err(err).Interface("content", content).Msg(message)
	return err
}

func logError(err error, message string, content interface{}) error {
	log.Error().Err(err).Interface("content", content).Msg(message)
	return err
}
