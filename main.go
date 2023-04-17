package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
)

var dockerSocket string
var apiKey string

func init() {
	dockerSocket = os.Getenv("DOCKER_SOCKET")
	if dockerSocket == "" {
		dockerSocket = "/var/run/docker.sock"
	}

	apiKey = os.Getenv("API_KEY")
	if apiKey == "" {
		panic("Warning: API_KEY environment variable is not set")
	}
}

func main() {
	http.HandleFunc("/ps", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || authHeader != fmt.Sprintf("Bearer %s", apiKey) {
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		containers, err := get("http://localhost/containers/json")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting running images: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(containers)
	})

	http.ListenAndServe(":8080", nil)
}

func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || authHeader != fmt.Sprintf("Bearer %s", apiKey) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func get(url string) ([]byte, error) {
	conn, err := net.Dial("unix", dockerSocket)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(_, _ string) (net.Conn, error) {
				return conn, nil
			},
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
