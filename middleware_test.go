package middleware

import (
	"fmt"
	"log"
	"net/http"
	"testing"
)

func TestStartServer(t *testing.T) {
	resource := []ResourceType{ResourceType{"name1", "GET", testHandler1(testHandler2(FinalHandler))}, ResourceType{"name2", "GET", testHandler2}}

	config := ConfigType{8080, "/fp", resource}
	StartServer(config)
}

func testHandler1(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Got a %s request for: %v", r.Method, r.URL)
		w.Write([]byte(fmt.Sprintf("Got a %s request for: %v", r.Method, r.URL)))
		log.Println("Handler finished processing request")
		return
	})
}

func testHandler2(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Got a %s request for: %v", r.Method, r.URL)
		w.Write([]byte(fmt.Sprintf("Got a %s request for: %v", r.Method, r.URL)))
		log.Println("Handler finished processing request")
		handler.ServeHTTP(w, r)
	})
}
