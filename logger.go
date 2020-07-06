package main

import "log"

var Logger *log.Logger

func logf(format string, v ...interface{}) {
	if Logger == nil {
		log.Printf(format, v...)
		return
	}
	Logger.Printf(format, v...)
}

func debugf(format string, v ...interface{}) {
	if *debug {
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
