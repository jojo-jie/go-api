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

// it's actually a good idea to implement 
// it's actually a good idea to implement a
// Flusher version for our writer as well

	*HttpWriter  // wrap our "normal" writer
	http.Flusher // keep a ref to the wrapped Flusher
	http.Flusher  // keep a ref to the wrapped Flusher
}

func (w *HttpWriterFlusher) Flush() {
	w.Flusher.Flush()
}

// modify the constructor to either return HttpWriter or
// HttpWriterFlusher depending on the writer being wrapped

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
