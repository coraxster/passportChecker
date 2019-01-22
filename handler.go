package passportChecker

import (
	"fmt"
	"github.com/go-chi/chi"
	"net/http"
)

type Handler struct {
	ch    ExistChecker
	chSql *SQLiteChecker
}

func MakeHandler(ch ExistChecker, chSql *SQLiteChecker) *Handler {
	return &Handler{ch: ch, chSql: chSql}
}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	value := chi.URLParam(r, "value")
	result, err := h.ch.Check([]interface{}{value})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_, err = w.Write([]byte(fmt.Sprintf("%v:%v", value, result[0])))
}

func (h *Handler) Count(w http.ResponseWriter, r *http.Request) {
	result, err := h.chSql.Count()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_, err = w.Write([]byte(fmt.Sprintf("count:%v", result)))
}

//func (h *Handler) GetFrom(w http.ResponseWriter, r *http.Request) {
//	ts := chi.URLParam(r, "ts")
//
//}
