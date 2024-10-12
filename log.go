package redisdb

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

type dWriter struct {
}

func (dr dWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	return dr.Write(p)
}

func (dr dWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write([]byte(time.Now().Format("2006-01-02 15:04:05") + " "))
	_, err = os.Stdout.Write(p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

var levelWriter dWriter = dWriter{}

var Logger = zerolog.New(levelWriter)

func Debug() *zerolog.Event {
	return Logger.Debug()
}
func Info() *zerolog.Event {
	return Logger.Info()
}
func Warn() *zerolog.Event {
	return Logger.Warn()
}

func Fatal() *zerolog.Event {
	return Logger.Fatal()
}
func Panic() *zerolog.Event {
	return Logger.Panic()
}
func Log() *zerolog.Event {
	return Logger.Log()
}
