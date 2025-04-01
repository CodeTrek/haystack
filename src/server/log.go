package server

import (
	"log"
	"os"
	"path/filepath"
	"search-indexer/conf"

	"gopkg.in/natefinch/lumberjack.v2"
)

func initLog() {
	if conf.Get().Server.LoggingStdout {
		log.SetOutput(os.Stdout)
	} else {
		dir := filepath.Join(conf.Get().Global.HomePath, "logs")
		logFile := filepath.Join(dir, "server.log")

		log.SetOutput(&lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    50, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   //days
			Compress:   true, // disabled by default
		})
	}
}
