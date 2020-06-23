package tlog

type Config struct {
	FileSize  int    `toml:"filesize" json:"filesize"`
	FileNum   int    `toml:"filenum" json:"filenum"`
	FileName  string `toml:"filename" json:"filename"`
	Level     string `toml:"level" json:"level"`
	Debug     bool   `toml:"debug" json:"debug"`
	Dir       string `toml:"dir" json:"dir"`
	UseSyslog bool   `toml:"use_syslog" json:"use_syslog"`
	SyslogTag string `toml:"syslog_tag" json:"syslog_tag"`
}

func (c *Config) check() {
	if c.FileSize == 0 {
		c.FileSize = 128
	}
	if c.FileNum == 0 {
		c.FileNum = 10
	}
	if c.FileName == "" {
		c.FileName = "INFO"
	}
	if c.Dir == "" {
		c.Dir = "./logs"
	}
	if c.Level == "" {
		c.Level = "DEBUG"
	}
}

func Init(c Config) {
	c.check()
	newLogger(c)
	l.run()
}

func Close() {
	l.stop()
}

func Debug(args ...interface{}) {
	l.p(DEBUG, args...)
}

func Debugf(format string, args ...interface{}) {
	l.pf(DEBUG, format, args...)
}

func Info(args ...interface{}) {
	l.p(INFO, args...)
}

func Infof(format string, args ...interface{}) {
	l.pf(INFO, format, args...)
}

func Warning(args ...interface{}) {
	l.p(WARNING, args...)
}

func Warningf(format string, args ...interface{}) {
	l.pf(WARNING, format, args...)
}

func Error(args ...interface{}) {
	l.p(ERROR, args...)
}

func Errorf(format string, args ...interface{}) {
	l.pf(ERROR, format, args...)
}

func Fatal(args ...interface{}) {
	l.p(FATAL, args...)
}

func Fatalf(format string, args ...interface{}) {
	l.pf(FATAL, format, args...)
}
