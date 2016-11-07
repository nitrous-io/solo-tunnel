package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Level uint8

const (
	ErrorLevel Level = iota
	WarnLevel
	InfoLevel
	DebugLevel
)

var LogLevel = WarnLevel
var ch = make(chan string, 1000)

func SetLogLevel(l string) {
	switch l {
	case "error":
		LogLevel = ErrorLevel
	case "warn":
		LogLevel = WarnLevel
	case "info":
		LogLevel = InfoLevel
	case "debug":
		LogLevel = DebugLevel
	}
}

func IsValidLogLevel(l string) bool {
	return l == "error" || l == "warn" || l == "info" || l == "debug"
}

func Start() {
	for {
		select {
		case msg := <-ch:
			fmt.Fprintln(os.Stdout, msg)
		}
	}
}

func send(msgType string, a ...interface{}) {
	ch <- fmt.Sprint(append([]interface{}{fmt.Sprintf("[nsync %s] ", msgType)}, a...)...)
}

func sendf(msgType, format string, a ...interface{}) {
	ch <- fmt.Sprintf(fmt.Sprintf("[nsync %s] %s", msgType, format), a...)
}

func Info(a ...interface{}) {
	if LogLevel >= InfoLevel {
		send("info", a...)
	}
}

func Infof(format string, a ...interface{}) {
	if LogLevel >= InfoLevel {
		sendf("info", format, a...)
	}
}

func Warn(a ...interface{}) {
	if LogLevel >= WarnLevel {
		send("warn", a...)
	}
}

func Warnf(format string, a ...interface{}) {
	if LogLevel >= WarnLevel {
		sendf("warn", format, a...)
	}
}

func Error(a ...interface{}) {
	if LogLevel >= ErrorLevel {
		send("error", a...)
	}
}

func Errorf(format string, a ...interface{}) {
	if LogLevel >= ErrorLevel {
		sendf("error", format, a...)
	}
}

func Debug(a ...interface{}) {
	if LogLevel >= DebugLevel {
		send("debug", a...)
	}
}

func Debugf(format string, a ...interface{}) {
	if LogLevel >= DebugLevel {
		sendf("debug", format, a...)
	}
}

func Fatal(a ...interface{}) {
	log.Fatal(append([]interface{}{"[nsync fatal] "}, a...))
}

func Fatalf(format string, a ...interface{}) {
	log.Fatalf(fmt.Sprintf("[nsync fatal] %s", format), a...)
}

func Panic(a ...interface{}) {
	log.Panic(append([]interface{}{"[nsync panic] "}, a...))
}

func Panicf(format string, a ...interface{}) {
	log.Panicf(fmt.Sprintf("[nsync panic] %s", format), a...)
}

func Event(name string, data interface{}) {
	message := struct {
		Name string      `json:"name"`
		Data interface{} `json:"data"`
	}{
		Name: name,
		Data: data,
	}

	b, err := json.Marshal(message)
	if err != nil {
		return
	}

	sendf("event", "%s", string(b))
}

type debugLogger struct{}

var debugLoggerSingleton = &debugLogger{}

func (d *debugLogger) Write(p []byte) (n int, err error) {
	Debug(string(bytes.TrimSpace(p)))
	return len(p), nil
}

func DebugLogger() *debugLogger {
	return debugLoggerSingleton
}
