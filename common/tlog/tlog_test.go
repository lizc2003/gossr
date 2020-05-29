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
		Tag:      "test",
		//Debug:     true,
		//SentryUrl: "http://d25c96785a184d6f8d051712a209a3a8:7f4a22c48897423caa9a0aec6bb93416@172.20.0.84:9000/5",
		UseSyslog: true,
		SyslogTag: "test",
	})
	for i := 0; i < 10; i++ {
		Info("xxxxxxxxxxasfsadjflasjfdlasjdfsajdfsadfjasjfdafjsfsa")
		Infof("%s", "xxxxxxxxxxasfsadjflasjfdlasjdfsajdfsadfjasjfdafjsfsa")
		InfoJson(map[string]interface{}{"test": time.Now().UnixNano()})
		Error("xxxxxxxxxxasfsadjflasjfdlasjdfsajdfsadfjasjfdafjsfsa")
		Errorf("%s", "xxxxxxxxxxasfsadjflasjfdlasjdfsajdfsadfjasjfdafjsfsa")
		ErrorJson(map[string]interface{}{"test": time.Now().UnixNano()})
	}
	time.Sleep(1 * time.Second)
}
