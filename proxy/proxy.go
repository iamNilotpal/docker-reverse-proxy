package proxy

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/iamNilotpal/drp/docker"
	"github.com/iamNilotpal/drp/server"
)

func CreateReverseProxy() http.Handler {
	router := chi.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			println("Forwarding request...")
			handleFunc(w, r)
		})
	})

	return router
}

func handleFunc(w http.ResponseWriter, r *http.Request) {
	hostname := r.Host
	subdomain := strings.Split(hostname, ".")[0]

	info, ok := docker.Get(subdomain)
	if !ok {
		server.Respond(w, http.StatusNotFound, "Not Found")
		return
	}

	port := strconv.Itoa(info.Port)
	target := "http://" + info.IpAddress + ":" + port

	parsedURL, _ := url.Parse(target)
	r.URL = parsedURL

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, "Could not reach origin server", 500)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.CopyN(w, resp.Body, resp.ContentLength)
}
