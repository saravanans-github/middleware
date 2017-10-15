package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	Name string
	// Only accepts on of the following: POST, GET, OPTIONS, CONNECT
	Method string
	// Handler function which will act on the incoming request
	Handler handlerFunc
}

// ErrorResponseType is a data structure that defines a HTTP error response from the Middleware
type ErrorResponseType struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type configError struct {
	Message string
}

// FinalHandler is a helper http.HandlerFunc as the final closure to the Resource handler
// Implementor can set his own final handler by config.FinalHandler = yourCustomFinalHandlerFunc
var FinalHandler http.HandlerFunc

// AllowedOrigins is an array of origins that are allowed if EnableCORS handler is used
var AllowedOrigins []string

// Local variables
type handlerFunc func(handler http.Handler) http.Handler

var gConfig ConfigType

// StartServer starts a server with the specified config
func StartServer(config ConfigType) error {

	// store a global reference of the config
	gConfig = config

	// first thing... ensure that the config is valid
	validateConfig(&config)

	// the implementor has an option to set his own final handler.
	if FinalHandler == nil {
		FinalHandler = http.HandlerFunc(final)
	}

	for i := 0; i < len(config.Resources); i++ {
		http.Handle(config.Path+config.Resources[i].Name, config.Resources[i].Handler(FinalHandler))
	}
	err := http.ListenAndServe(":"+fmt.Sprint(config.Port), nil)
	if err != nil {
		log.Panic(err)
	}

	return nil
}

// Pre-built Handlers that can be used.

// IsRequestValid validates the request for a particular resource
func IsRequestValid(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a supported Method
		method, err := getMethodByResourceName(r.URL.EscapedPath())
		if err != nil {
			message, status := getErrorResponse(500, err.Error())
			http.Error(w, message, status)
			return
		}

		if r.Method != method {
			message, status := getErrorResponse(400, "Method not supported.")
			http.Error(w, message, status)
			return
		}

		// Check for a request body
		if r.ContentLength == 0 {
			message, status := getErrorResponse(400, "Body is empty.")
			http.Error(w, message, status)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// EnableCORS enables Cross Origin Resource Sharing for a particular resource
func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check if it is a allowed Origin
		if !isAllowedOrigin(r.Header.Get("Origin")) {
			log.Printf("Origin %s not allowed", r.Header.Get("Origin"))
			http.Error(w, "This origin is not authorised to access", 401)

			return
		}

		// Get the method configured for this Resource
		method, err := getMethodByResourceName(r.URL.EscapedPath())
		if err != nil {
			log.Fatal(err)

			return
		}

		log.Println("Setting CORS headers")
		w.Header().Set("Access-Control-Allow-Methods", method)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		w.Header().Set("Content-Type", "text/html")

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
		return &configError{"No valid PORT set."}
	}

	if (*config).Path == "" {
		return &configError{"No valid PATH set."}
	}

	if len((*config).Resources) == 0 {
		return nil
	}

	for i := 0; i < len(config.Resources); i++ {
		if (*config).Resources[i].Name == "" {
			return &configError{"No valid RESOURCE set for path [" + (*config).Path + "]"}
		}

		if (*config).Resources[i].Method == "" {
			return &configError{"No valid METHOD set for resource [" + (*config).Resources[i].Name + "]"}
		}

		if (*config).Resources[i].Handler == nil {
			return &configError{"No valid HANDLER set for resource [" + (*config).Resources[i].Name + "]"}
		}
	}

	return nil
}

func getMethodByResourceName(name string) (string, error) {
	for _, a := range gConfig.Resources {
		if gConfig.Path+a.Name == name {
			return a.Method, nil
		}
	}

	return "", &configError{"Resource with name " + name + " not found."}
}

func isAllowedOrigin(origin string) bool {
	for _, allowed := range AllowedOrigins {
		if origin == allowed {
			return true
		}
	}

	// no valid origin found; so return false
	return false
}

func (e *configError) Error() string {
	return e.Message
}

func getErrorResponse(status int, message string) (string, int) {
	resBody, err := json.Marshal(ErrorResponseType{status, message})
	if err != nil {
		log.Fatal(err)
		return "Server error.", 500
	}

	return string(resBody), status
}
