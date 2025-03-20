package validation

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		// we check that client can receive compressed data in gzip format from server
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// wrap the original http.ResponseWriter with a new one with compression support
			cw := NewCompressWriter(w)
			// change the original http.ResponseWriter to a new one
			ow = cw
			// send all compressed data to client after middleware completes
			defer cw.Close()
		}

		// check that the client sent server compressed data in gzip format
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// wrap the request body in io.Reader with decompression support
			cr, err := NewCompressReader(r.Body)
			if err != nil {
				log.Printf("Error decompressing request body: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// swaps body
			r.Body = cr
			defer cr.Close()
		}
		// return control to handler
		h.ServeHTTP(ow, r)
	})
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		gz, _ := gzip.NewWriterLevel(nil, gzip.DefaultCompression)
		return gz
	},
}

type optimizedGzipResponseWriter struct {
	http.ResponseWriter
	gz             *gzip.Writer
	buffer         *bytes.Buffer
	minSizeToFlush int
	headerSent     bool
	statusCode     int
}

func OptimizedGzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		if isSmallRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		buffer := bufferPool.Get().(*bytes.Buffer)
		buffer.Reset()
		defer bufferPool.Put(buffer)

		gz := gzipWriterPool.Get().(*gzip.Writer)
		gz.Reset(buffer)
		defer gzipWriterPool.Put(gz)

		gzipResponseWriter := &optimizedGzipResponseWriter{
			ResponseWriter: w,
			gz:             gz,
			buffer:         buffer,
			minSizeToFlush: 4096,
			headerSent:     false,
		}
		defer gzipResponseWriter.finalize()
		next.ServeHTTP(gzipResponseWriter, r)
	})
}

func isSmallRequest(r *http.Request) bool {
	length := r.Header.Get("Content-Length")
	if length == "" {
		return false
	}
	size, err := strconv.Atoi(length)
	if err != nil || size <= 0 {
		return false
	}
	if size <= 1024 {
		return true
	}
	return false
}

func (w *optimizedGzipResponseWriter) Write(p []byte) (int, error) {
	if !w.headerSent {
		// Устанавливаем заголовки для gzip только перед первой записью
		w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		w.ResponseWriter.Header().Add("Vary", "Accept-Encoding")

		// Если код статуса не установлен, используем 200 OK
		if w.statusCode != 0 {
			w.ResponseWriter.WriteHeader(w.statusCode)
		} else {
			w.ResponseWriter.WriteHeader(http.StatusOK)
		}

		w.headerSent = true
	}

	// Пишем данные в gzip writer
	n, err := w.gz.Write(p)
	if err != nil {
		return n, err
	}

	// Выполняем Flush только если размер буфера превысил минимальный
	if w.buffer.Len() >= w.minSizeToFlush {
		w.flushBuffer()
	}

	return n, nil
}

func (w *optimizedGzipResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *optimizedGzipResponseWriter) flushBuffer() {
	w.gz.Flush()
	w.buffer.WriteTo(w.ResponseWriter)
	w.buffer.Reset()
}

func (w *optimizedGzipResponseWriter) finalize() {
	if w.gz == nil {
		return
	}

	if !w.headerSent && w.buffer.Len() > 0 {
		w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		w.ResponseWriter.Header().Add("Vary", "Accept-Encoding")
		if w.statusCode != 0 {
			w.ResponseWriter.WriteHeader(w.statusCode)
		} else {
			w.ResponseWriter.WriteHeader(http.StatusOK)
		}
	}

	w.gz.Close()
	if w.buffer.Len() > 0 {
		w.buffer.WriteTo(w.ResponseWriter)
	}
	w.gz.Reset(io.Discard)
}

func (m *MiddlewareController) OptimizedValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		if !needsValidation(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func needsValidation(path string) bool {
	return strings.Contains(path, "/update") ||
		strings.Contains(path, "/updates") ||
		strings.Contains(path, "/value")
}
