package validation

import (
	"log"
	"net/http"
	"strings"
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
