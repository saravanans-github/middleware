package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
)

// ConfigType is a required to start server
type ConfigType struct {
	Port      uint
	Path      string
	Resources []ResourceType
}

// ResourceType is a data structure for endpoints that is defined for Config.Path
type ResourceType struct {
	// Name of the Resource. This value has to escaped for special characters
	Path string
	// Only accepts on of the following: POST, GET, OPTIONS, CONNECT
	Method string
	// Handler function which will act on the incoming request
	Handler http.Handler
}

const _ErrorInvalidPath = "No valid PATH set"
const _ErrorInvalidPort = "No valid PORT set"
const _ErrorEmptyResources = "No resources were set for path: "    // + config.path
const _ErrorInvalidMethod = "No valid Method set for resource: "   // + config.resources[i].path
const _ErrorInvalidHandler = "No valid Handler set for resource: " // + config.resources[i].handler
const _ErrorResourceNotFound = "No such resource was found: "      // + config.resources[i].path
const _HTTPErrorUnauthorisedOrigin = "This origin is not authorised to access."

type errorResponseType struct {
	Status  int    `json:"status"` // default status is 0
	Message string `json:"message"`
}

// FinalHandler is a helper http.HandlerFunc as the final closure to the Resource handler
// Implementor can set his own final handler by config.FinalHandler = yourCustomFinalHandlerFunc
var FinalHandler http.HandlerFunc

// AllowedOrigins is an array of origins that are allowed if EnableCORS handler is used
var AllowedOrigins []string

var _config ConfigType
var _server http.Server
var _mux map[string]http.Handler

// StartServer starts a server with the specified config
func StartServer(config ConfigType) (err error) {

	// store a global reference of the config
	_config = config

	// first thing... ensure that the config is valid
	if err = validateConfig(&config); err != nil {
		log.Println(err.Error())
		return
	}

	// the implementor has an option to set his own final handler.
	if FinalHandler == nil {
		FinalHandler = http.HandlerFunc(final)
	}

	//
	_mux = make(map[string]http.Handler)

	for i := 0; i < len(config.Resources); i++ {
		_mux[config.Path+config.Resources[i].Path] = config.Resources[i].Handler
	}

	_server = http.Server{Addr: (":" + fmt.Sprint(config.Port)), Handler: http.HandlerFunc(serve)}

	//	go func() {
	if err = _server.ListenAndServe(); err != nil {
		// cannot panic, because this probably is an intentional close
	}
	//	}()

	return nil
}

// StopServer stops a running server
func StopServer() error {
	if err := _server.Shutdown(nil); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}

	return nil
}

// Pre-built Handlers that can be used.

// IsRequestValid validates the request for a particular resource
func IsRequestValid(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Starting Is Request Valid check...")

		// Check if it's a supported Method
		log.Printf("	Finding allowed method for resource: %s...", r.URL.EscapedPath())
		method, err := getMethodByResourceName(r.URL.EscapedPath())
		if err != nil {
			message, status := GetErrorResponse(500, err.Error())
			http.Error(w, message, status)
			return
		}
		log.Printf("	Finding allowed method for resource: %s... FOUND [%s]", r.URL.EscapedPath(), method)
		log.Printf("	Incoming Method: %s", r.Method)

		if r.Method != method {
			log.Println("Start Is Request Valid check... FAILED")
			message, status := GetErrorResponse(400, "Method not supported.")
			http.Error(w, message, status)
			return
		}

		// Check for a request body
		if r.ContentLength == 0 {
			message, status := GetErrorResponse(400, "Body is empty.")
			http.Error(w, message, status)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// EnableCORS enables Cross Origin Resource Sharing for a particular resource
func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Setting CORS headers... ")

		// Check if it is a allowed Origin
		log.Println("	Validating request Origin... ")
		if err := isAllowedOrigin(r.Header.Get("Origin")); err != nil {
			log.Printf("	Validating request Origin... FAILED. Origin [%s] not found/allowed", r.Header.Get("Origin"))
			log.Println("Setting CORS headers... ABORTED")

			message, status := GetErrorResponse(401, err.Error())
			http.Error(w, message, status)
			return
		}

		// Get the method configured for this Resource
		log.Println("	Validating request Method... ")
		method, err := getMethodByResourceName(r.URL.EscapedPath())
		if err != nil {
			log.Printf("	Validating request Method... FAILED. %s", err.Error())
			log.Println("Setting CORS headers... ABORTED")

			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Access-Control-Allow-Methods", method)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		log.Println("Setting CORS headers... DONE")

		// Stop here if its Preflighted OPTIONS request
		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Internal helper functions

func final(w http.ResponseWriter, r *http.Request) {
	log.Println("Executing finalHandler")
}

func validateConfig(config *ConfigType) error {
	// Check if the port is a valid port
	if (*config).Port == 0 || (*config).Port > 65536 {
		return errorResponseType{Message: _ErrorInvalidPort}
	}

	if (*config).Path == "" || string((*config).Path[0]) != "/" {
		return errorResponseType{Message: _ErrorInvalidPath}
	}

	if len((*config).Resources) == 0 {
		return errorResponseType{Message: _ErrorEmptyResources + (*config).Path}
	}

	for i := 0; i < len(config.Resources); i++ {
		if (*config).Resources[i].Path == "" || string((*config).Path[0]) != "/" {
			return errorResponseType{Message: _ErrorInvalidPath + (*config).Path}
		}

		if (*config).Resources[i].Method == "" {
			return errorResponseType{Message: _ErrorInvalidMethod + (*config).Resources[i].Path}
		}

		if (*config).Resources[i].Handler == nil {
			return errorResponseType{Message: _ErrorInvalidHandler}
		}
	}

	return nil
}

func getMethodByResourceName(path string) (string, error) {
	for _, a := range _config.Resources {
		if _config.Path+a.Path == path {
			return a.Method, nil
		}
	}

	return "", errorResponseType{Status: 400, Message: _ErrorResourceNotFound + path}
}

func isAllowedOrigin(origin string) error {
	for _, allowed := range AllowedOrigins {
		if origin == allowed {
			return nil
		}
	}

	// no valid origin found; so return false
	return errorResponseType{Message: _HTTPErrorUnauthorisedOrigin}
}

func (e errorResponseType) Error() string {
	return e.Message
}

func GetErrorResponse(status int, message string) (string, int) {
	resBody, err := json.Marshal(errorResponseType{status, message})
	if err != nil {
		log.Fatal(err)
		return "Server error.", 500
	}

	return string(resBody), status
}

func serve(w http.ResponseWriter, r *http.Request) {
	log.Println("Finding handler...")
	if h, ok := _mux[r.URL.String()]; ok {
		log.Printf("Finding handler... FOUND [%s]", runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name())
		h.ServeHTTP(w, r)
		return
	}

	log.Println("Finding handler... NOT found")
	// If we can't find a handler - Give throw a Resource Not Found error
	resourceNotFound(w)
}

func resourceNotFound(w http.ResponseWriter) {
	message, status := GetErrorResponse(404, _ErrorResourceNotFound+"Check that you entered your URL correctly")
	http.Error(w, message, status)
}
