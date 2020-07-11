package pkg

import "log"

var Logger *log.Logger

var DebugMode = false

func logf(format string, v ...interface{}) {
	if Logger == nil {
		log.Printf(format, v...)
		return
	}
	Logger.Printf(format, v...)
}

func debugf(format string, v ...interface{}) {
	if DebugMode {
		logf("[DEBUG] "+format, v)
	}
}

func fatalf(format string, v ...interface{}) {
	if Logger == nil {
		log.Fatalf(format, v...)
		return
	}
	Logger.Fatalf(format, v...)
}
