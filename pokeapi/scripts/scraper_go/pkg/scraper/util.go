package scraper

import (
	"regexp"
	"strconv"
)

const (
	ERRID = 20000
)

func findSubmatchID(key string, value []byte, parser map[string]*regexp.Regexp) int64 {
	matches := parser[key].FindSubmatch(value)
	if len(matches) < 1 || string(matches[1]) == "" {
		return ERRID
	}
	id, err := strconv.ParseInt(string(matches[1]), 10, 64)
	if err != nil {
		return ERRID
	}
	return id
}

func findSubmatch(key string, value []byte, parser map[string]*regexp.Regexp) string {
	matches := parser[key].FindSubmatch(value)
	if len(matches) < 1 {
		return ""
	}
	return string(matches[1])
}

func findSubmatchRound(key string, value []byte, parser map[string]*regexp.Regexp) string {
	matches := parser[key].FindSubmatch(value)
	if len(matches) < 1 {
		return ""
	}
	res, err := strconv.ParseFloat(string(matches[1]), 64)
	if err != nil {
		return ""
	}

	return strconv.Itoa(int(res * 10))
}
