package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func StringToInt64(s string, defaultVal int64) int64 {
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n
	}
	return defaultVal
}

func FormatFullTime(t time.Time) string {
	y, m, d := t.Date()
	return fmt.Sprintf("%4d-%02d-%02d %02d:%02d:%02d", y, m, d, t.Hour(), t.Minute(), t.Second())
}

func GetDomainFromHost(host string) string {
	pos := strings.Index(host, ":")
	if pos >= 0 {
		host = host[:pos]
	}
	parts := strings.Split(host, ".")
	sz := len(parts)
	if sz <= 2 {
		return host
	}

	partsLen := 2
	part2 := parts[sz-2]
	switch part2 {
	case "com":
		fallthrough
	case "org":
		fallthrough
	case "net":
		fallthrough
	case "edu":
		fallthrough
	case "gov":
		partsLen = 3
	default:
		part1 := parts[sz-1]
		if part1 == "uk" || part1 == "jp" {
			switch part2 {
			case "co":
				fallthrough
			case "ac":
				fallthrough
			case "me":
				partsLen = 3
			}
		}
	}
	return strings.Join(parts[sz-partsLen:], ".")
}
