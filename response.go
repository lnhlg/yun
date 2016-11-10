package yun

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

const (
	noWritten = -1
)

type (
	//ResponseWriter 响应写接口
	ResponseWriter interface {
		http.ResponseWriter
		http.Hijacker
		http.Flusher
		http.CloseNotifier

		//获取当前请求的响应状态码
		Status() int

		// 返回已写入HTTP响应体的字节数
		Size() int

		// 如果响应正文已被写入，则返回True
		Written() bool
	}

	responseWriter struct {
		http.ResponseWriter
		status int
		size   int
	}
)

func (w *responseWriter) reset(wr http.ResponseWriter) {
	w.ResponseWriter = wr
	w.status = http.StatusOK
	w.size = noWritten
}

// 写HTTP信息头 (status code + headers).
func (w *responseWriter) WriteHeader(code int) {
	if !w.Written() {
		w.status = code
		w.size = 0
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.writeHeader()
	n, err := w.ResponseWriter.Write(data)
	w.size += n
	return n, err
}

func (w *responseWriter) Written() bool {
	return w.size != noWritten
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) writeHeader() {
	w.WriteHeader(w.status)
}

// 实现http.Hijacker接口
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.size < 0 {
		w.size = 0
	}

	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

// 实现http.CloseNotify接口
func (w *responseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// 实现http.Flush接口
func (w *responseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}
