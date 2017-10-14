package middleware

import (
	"fmt"
	"log"
	"net/http"
	"testing"
)

var config ConfigType

func TestStartServer(t *testing.T) {
	resource := []ResourceType{ResourceType{"name1", "GET", testHandler}, ResourceType{"name2", "GET", testHandler2}}

	config = ConfigType{8080, "/fp", resource}
	StartServer(config)
}

func testHandler(handler http.Handler) http.Handler {
	return testHandler1(testHandler2(FinalHandler))
}

func testHandler1(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Got a %s request for: %v", r.Method, r.URL)
		w.Write([]byte(fmt.Sprintf("\nHandler 1 - Got a %s request for: %v", r.Method, r.URL)))
		log.Println("Handler finished processing request")
		next.ServeHTTP(w, r)
	})
}

func testHandler2(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Got a %s request for: %v", r.Method, r.URL)
		w.Write([]byte(fmt.Sprintf("\nHandler 2 - Got a %s request for: %v", r.Method, r.URL)))
		log.Println("Handler finished processing request")
	})
}
