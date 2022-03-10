package downloads

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Header struct {
	ContentType   string
	ContentLength int
	FileName      string
}

func Head(url string) (header Header, err error) {
	res, err := http.Head(url)
	if err != nil {
		return
	}
	len, err := parseContentLength(res.Header)
	if err != nil {
		return
	}
	name, err := parseFileName(res.Header)
	mime := strings.Split(res.Header.Get("Content-Type"), ";")[0]
	header = Header{
		mime,
		len,
		name,
	}
	return
}

func parseContentLength(h http.Header) (int, error) {
	val := h.Get("Content-Length")
	return strconv.Atoi(val)
}

func parseFileName(h http.Header) (name string, err error) {
	val := h.Get("Content-Disposition")
	split := strings.Split(val, "filename=")
	if len(split) < 2 {
		err = errors.New("file name not found")
		return
	}
	name = split[1]
	if strings.HasPrefix(name, "\"") {
		name = strings.Trim(name, "\"")
	}
	return
}
