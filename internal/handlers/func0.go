package handlers

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"gomuncool/internal/models"
)

func Cap(rwr http.ResponseWriter, req *http.Request) {

	status := http.StatusOK
	ip := getIP(req)
	rwr.WriteHeader(http.StatusOK)
	fmt.Fprintf(rwr, `{"status":%d, "IP":"%s"}`, status, ip)

}
func getIP(r *http.Request) string {
	// Получаем IP из заголовков, если есть прокси
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For может содержать список IP через запятую
		ips := strings.Split(ip, ",")
		ip = strings.TrimSpace(ips[0])
		return ip
	}

	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Если нет заголовков, получаем из RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

func GetUser(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(req)
	userName := vars["userName"]

	role, err := models.DataBase.GetUser(req.Context(), userName)
	if err != nil {
		rwr.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rwr, `{"wrong user name":"%s"}`, userName)
		return
	}
	rwr.WriteHeader(http.StatusOK)
	fmt.Fprintf(rwr, `{"user":"%s", "role":"%s"}`, userName, role)

}

func PutUser(rwr http.ResponseWriter, req *http.Request) {

	rwr.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(req)
	userName := vars["userName"]
	role := vars["role"]

	if userName == "" || role == "" {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	err := models.DataBase.PutUser(req.Context(), userName, role)
	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rwr, `{"BAD user":"%s", "err":"%v"}`, userName, err)
		return
	}
	rwr.WriteHeader(http.StatusOK)
	fmt.Fprintf(rwr, `{"user added":"%s", "with role":"%s"}`, userName, role)

}
