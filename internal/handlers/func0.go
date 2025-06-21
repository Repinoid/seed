package handlers

import (
	"fmt"
	"net"
	"net/http"
	"strings"
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
