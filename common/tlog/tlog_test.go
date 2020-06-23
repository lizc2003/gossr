package tlog

import (
	"testing"
	"time"
)

func TestTlog(t *testing.T) {
	Init(Config{
		FileSize: 128,
		FileNum:  5,
		FileName: "test",
		Level:    "DEBUG",
		Dir:      "./logs",
		//Debug:     true,
		UseSyslog: true,
		SyslogTag: "test",
	})

	for i := 0; i < 100; i++ {
		go doLog()
	}
	time.Sleep(5 * time.Second)

	Close()
}

func doLog() {
	var a []byte
	var b map[byte]byte
	Info("xxxxxxxxxxasfsadjflasjfdlasjdfsajdfsadfjasjfdafjsfsa")
	Infof("%s", "xxxxxxxxxxasfsadjflasjfdlasjdfsajdfsadfjasjfdafjsfsa")
	Error(a)
	Errorf("%s", a)
	Error(b)
	Errorf("%v", b)
}
