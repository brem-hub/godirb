package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type loggerManager struct {
	url string
	lg  loggerCust
}

func (lm *loggerManager) New(url string) {
	lm.url = url
}
func (lm *loggerManager) Start() {
	lm.lg.createLogger(lm.url)
}
func (lm *loggerManager) Close() {
	lm.lg.closeLogger()
}
func (lm *loggerManager) GetLogger() loggerCust {
	return lm.lg
}

type loggerCust struct {
	mutex  sync.Mutex
	logger *log.Logger
	file   *os.File
}

func (l *loggerCust) Println(message string) {
	l.logger.Printf("[%02d:%02d:%02d] %s\n", time.Now().Hour(), time.Now().Minute(), time.Now().Second(), message)
}

func (l *loggerCust) createLogger(url string) {
	path := clearURL(url)
	timer := time.Now()
	file, err := os.Create(fmt.Sprintf("../log/log_%s_%d-%02d-%02d_%02d-%02d-%02d", path, timer.Year(), timer.Month(),
		timer.Day(), timer.Hour(), timer.Minute(), timer.Second()))
	if err != nil {
		fmt.Println(err)
		return
	}
	l.file = file
	log := log.New(file, "", 0)
	l.logger = log
	l.mutex = sync.Mutex{}
	l.Println("Logger is started")
}

func (l *loggerCust) closeLogger() {
	l.Println("Logger is closed")
	l.file.Close()
}

func clearURL(url string) string {
	var path string
	if strings.Contains(url, "127.0.0.1") {
		path = "local"
	} else {
		switch {
		case strings.Contains(url, ".com"):
			path = strings.Replace(url, ".com", "", -1)
		case strings.Contains(url, ".ru"):
			path = strings.Replace(url, ".ru", "", -1)
		}
		switch {
		case strings.Contains(path, "https://"):
			path = strings.Replace(path, "https://", "", -1)
		case strings.Contains(path, "http://"):
			path = strings.Replace(path, "http://", "", -1)
		}
	}
	return path
}
