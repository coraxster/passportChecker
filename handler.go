package passportChecker

import (
	"fmt"
	"github.com/go-chi/chi"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	ch    ExistChecker
	chSql *MySQLChecker
}

func MakeHandler(ch ExistChecker, chSql *MySQLChecker) *Handler {
	return &Handler{ch: ch, chSql: chSql}
}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	value := chi.URLParam(r, "value")
	p, err := MakePassport(strings.Replace(strings.Replace(value, " ", "", -1), "-", "", -1), "")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	result, err := h.ch.Check([]interface{}{p.Uint64()})
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

func (h *Handler) Count(w http.ResponseWriter, r *http.Request) {
	result, err := h.chSql.Count()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_, err = w.Write([]byte(fmt.Sprintf("count:%v", result)))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (h *Handler) GetFrom(w http.ResponseWriter, r *http.Request) {
	ts, err := strconv.ParseInt(chi.URLParam(r, "ts"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Println(ts)
	result, err := h.chSql.GetFrom(time.Unix(ts, 0))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	for _, v := range result {
		_, err = w.Write([]byte(fmt.Sprintf("%v\r\n", v)))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}
