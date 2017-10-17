package middleware

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"testing"
)

var config ConfigType

func initTest() {
	log.SetFlags(0)
}

func TestStartStopServer(t *testing.T) {
	initTest()

	resource := []ResourceType{ResourceType{"/name1", "POST", testStartStopServerHandler(FinalHandler)}}

	config = ConfigType{Port: 8080, Path: "/fp", Resources: resource}
	StartServer(config)

	if status := sendRequest("http://localhost:8080/fp/name1", []byte("Hello World")); status != 200 {
		StopServer()
		t.FailNow()
	} else {
		StopServer()
	}
}

func TestIsRequestValid_Negative_WrongMethod(t *testing.T) {
	initTest()

	resource := []ResourceType{ResourceType{"/name1", "GET", IsRequestValid(testHandler2(FinalHandler))}}

	config = ConfigType{Port: 8080, Path: "/fp", Resources: resource}
	StartServer(config)
	if status := sendRequest("http://localhost:8080/fp/name1", []byte("Hello World")); status != 400 {
		StopServer()
		t.FailNow()
	} else {
		StopServer()
	}
}

func TestEnableCORS_Negative_NoHeader(t *testing.T) {
	initTest()

	resource := []ResourceType{ResourceType{"/name1", "GET", EnableCORS(testHandler2(FinalHandler))}}

	config = ConfigType{Port: 8080, Path: "/fp", Resources: resource}
	StartServer(config)
	if status := sendRequest("http://localhost:8080/fp/name1", []byte("Hello World")); status != 401 {
		StopServer()
		t.FailNow()
	} else {
		StopServer()
	}
}

func TestValidateConfig_Negative_ForInvalidPath(t *testing.T) {
	resource := []ResourceType{ResourceType{"/name1", "POST", IsRequestValid(testHandler2(FinalHandler))}}

	config = ConfigType{Port: 8080, Path: "", Resources: resource}
	if err := StartServer(config); !strings.HasPrefix(err.Error()[:], _ErrorInvalidPath) {
		t.FailNow()
	}

	config = ConfigType{Port: 8080, Path: "fp", Resources: resource}
	if err := StartServer(config); !strings.HasPrefix(err.Error()[:], _ErrorInvalidPath) {
		t.FailNow()
	}
}

func TestValidateConfig_Negative_ForInvalidPort(t *testing.T) {
	resource := []ResourceType{ResourceType{"/name1", "POST", IsRequestValid(testHandler2(FinalHandler))}}

	config = ConfigType{Port: 1000000, Path: "/fp", Resources: resource}
	if err := StartServer(config); !strings.HasPrefix(err.Error()[:], _ErrorInvalidPort) {
		t.FailNow()
	}
}

func TestValidateConfig_Negative_ForEmptyResources(t *testing.T) {
	resource := []ResourceType{}

	config = ConfigType{Port: 8080, Path: "/fp", Resources: resource}
	if err := StartServer(config); !strings.HasPrefix(err.Error()[:], _ErrorEmptyResources) {
		t.FailNow()
	}
}

func testStartStopServerHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}

func testIsRequestValidHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func testHandler2(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func testResourceNotFoundHandler(w http.ResponseWriter, r *http.Request) {
}

func sendRequest(url string, data []byte) int {

	res, err := http.Post(url, "application/json", bytes.NewBuffer(data)) //http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	return res.StatusCode
}
