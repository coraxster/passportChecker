package passportChecker

import (
	"fmt"
	"github.com/go-chi/chi"
	"net/http"
)

type Handler struct {
	ch ExistChecker
}

func MakeHandler(ch ExistChecker) *Handler {
	return &Handler{ch: ch}
}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	value := chi.URLParam(r, "value")
	result, err := h.ch.Check([]string{value})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_, err = w.Write([]byte(fmt.Sprintf("%v:%v", value, result[0])))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
