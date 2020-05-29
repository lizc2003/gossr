package tlog

type LEVEL byte

const (
	ALL LEVEL = iota
	DEBUG
	INFO
	WARNING
	ERROR
	FATAL
)

var levelText = map[LEVEL]string{
	ALL:     "ALL",
	DEBUG:   "DEBUG",
	INFO:    "INFO",
	WARNING: "WARNING",
	ERROR:   "ERROR",
	FATAL:   "FATAL",
}

func getLevel(level string) LEVEL {
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARNING":
		return WARNING
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return ALL
	}
}
