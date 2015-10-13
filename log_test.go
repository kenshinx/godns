package main

import (
	"bufio"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConsole(t *testing.T) {
	logger := NewLogger()
	logger.SetLogger("console", nil)
	logger.SetLevel(LevelInfo)

	logger.Debug("debug")
	logger.Info("info")
	logger.Notice("notiece")
	logger.Warn("warn")
	logger.Error("error")
}

func TestFile(t *testing.T) {
	logger := NewLogger()
	logger.SetLogger("file", map[string]interface{}{"file": "test.log"})
	logger.SetLevel(LevelInfo)

	logger.Debug("debug")
	logger.Info("info")
	logger.Notice("notiece")
	logger.Warn("warn")
	logger.Error("error")

	time.Sleep(time.Second)

	f, err := os.Open("test.log")
	if err != nil {
		t.Fatal(err)
	}
	b := bufio.NewReader(f)
	linenum := 0
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			break
		}
		if len(line) > 0 {
			linenum++
		}
	}

	Convey("Test Log File Handler", t, func() {
		Convey("file line nums should be 4", func() {
			So(linenum, ShouldEqual, 4)
		})
	})

	os.Remove("test.log")
}
