package middleware

import (
	"fmt"
	"html"
	"log"
	"net/http"
)

// ConfigType is a required to start server
type ConfigType struct {
	Port     uint
	Path     string
	Resource []ResourceType
}

type handlerFunc func(handler http.Handler) http.Handler

// ResourceType is an endpoint that is accesible under Config.Path
type ResourceType struct {
	Name    string
	Method  string
	Handler handlerFunc
}

type configError struct {
	Message string
}

var FinalHandler http.HandlerFunc

// StartServer starts a server with the specified config
func StartServer(config ConfigType) error {
	// first thing... ensure that the config is valid
	validateConfig(&config)

	FinalHandler = http.HandlerFunc(final)
	http.Handle(config.Path+"/"+config.Resource[0].Name, start(config.Resource[0].Handler(FinalHandler)))
	http.Handle(config.Path+"/"+config.Resource[1].Name, start(config.Resource[1].Handler(FinalHandler)))
	err := http.ListenAndServe(":"+fmt.Sprint(config.Port), nil)
	if err != nil {
		log.Panic(err)
	}

	return nil
}

func start(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Executing startHandler")
		next.ServeHTTP(w, r)
	})
}

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

	// if (*config).Resource.Name == "" {
	// 	return &configError{"No valid RESOURCE set for path [" + (*config).Path + "]"}
	// }

	// if (*config).Resource.Method == "" {
	// 	return &configError{"No valid METHOD set for resource [" + (*config).Resource.Name + "]"}
	// }

	// if (*config).Resource.Handler == nil {
	// 	return &configError{"No valid HANDLER set for resource [" + (*config).Resource.Name + "]"}
	// }

	return nil
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Setting CORS headers")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Content-Type", "text/html")

		// Stop here if its Preflighted OPTIONS request
		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (e *configError) Error() string {
	return e.Message
}

func wvHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from WV, %q, %q", html.EscapeString(r.URL.Path), r.URL.Query())
}

func fpHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from FP, %q, %q", html.EscapeString(r.URL.Path), r.URL.Query())
}
