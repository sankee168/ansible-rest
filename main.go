package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

func main() {

	// Initialize port
	port := ":3000"
	envPort := os.Getenv("ASNIBLE_REST_PORT")

	if envPort != "" {
		port = ":" + envPort
		log.Infof("starting ansible-rest on port: %s", port)
	}

	// Start the router
	r := mux.NewRouter()

	//attaching requestId in logs
	r.Use(CorrelationMiddleware)
	// route /ansibletasks to ansible executor
	r.HandleFunc("/ansibletasks", AnsibleTaskHandler).Methods("POST")
	log.Infof("starting ansible-rest on port: %s", port)
	corsObj := handlers.AllowedOrigins([]string{"*"})
	// Bind to a port and pass our router in
	log.Fatalln(http.ListenAndServe(port, handlers.CORS(corsObj)(r)))
}

// Set the logger
func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	file, err := os.OpenFile("ansible-rest.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(file)
	} else {
		log.Infof("Failed to log into a file, using defalt stderr")
	}
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	// log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

//CorrelationMiddleware is for attaching request id in all the logs from ansible
func CorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestID := uuid.NewV4()
		entry := log.WithFields(log.Fields{
			"requestID": requestID,
		})
		ctx := context.WithValue(r.Context(), "RequestLogger", entry)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
