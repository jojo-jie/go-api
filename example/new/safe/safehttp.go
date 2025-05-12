package safe

import "net/http"

type HttpWriter struct {
	w http.ResponseWriter // wrap an existing writer
}

func (w *HttpWriter) Header() http.Header {
	return w.w.Header()
}

func (w *HttpWriter) Write(data []byte) (int, error) {
	return w.w.Write(data)
}

func (w *HttpWriter) WriteHeader(statusCode int) {
	w.w.WriteHeader(statusCode)
}

type HttpWriterFlusher struct {
	*HttpWriter
	http.Flusher
}

func (w *HttpWriterFlusher) Flush() {
	w.Flusher.Flush()
}

func NewHttpWriter(w http.ResponseWriter) http.ResponseWriter {
	httpWriter := &HttpWriter{
		w: w,
	}

	if flusher, ok := w.(http.Flusher); ok {
		return &HttpWriterFlusher{
			HttpWriter: httpWriter,
			Flusher:    flusher,
		}
	}

	return httpWriter
}
