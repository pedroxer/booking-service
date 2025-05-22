package prometheus

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type httpHandler struct {
}

func (h httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
func RunRestServer() error {
	http.Handle("/metrics", promhttp.Handler())
	h1 := httpHandler{}
	http.Handle("/health", h1)
	err := http.ListenAndServe(":2121", nil)
	return err
}
