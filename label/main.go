// app.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"frm/label/storage"
	"log"
	"net/http"
	"path/filepath"
	"time"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/gorilla/mux"
)

// App export
type App struct {
	Router  *mux.Router
	Storage *storage.Storage
}

// Request body
type Body struct {
	Bucket string
	Object string
	Output string
}

func (app *App) labelHandler(w http.ResponseWriter, req *http.Request) {
	//READ BODY
	var b Body
	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(req.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	uri := "gs://" + b.Bucket + "/" + b.Object
	//get image annotations
	image := vision.NewImageFromURI(uri)
	annotations, err := client.DetectLabels(ctx, image, nil, 5)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fileName := filepath.Base(b.Object)
	s := fmt.Sprintf(`{ "object": "%s", "bucket": "%s" }`, "No Label", b.Output)
	//Build a file path for each label detected, then save it to GCS
	if len(annotations) > 0 {
		var target string
		for _, annotation := range annotations {
			target = target + "/" + annotation.Description + "/" + fileName
		}
		target = target + "/" + fileName
		app.Storage.CopyFile(b.Bucket, b.Object, b.Output, target)
		s = fmt.Sprintf(`{ "object": "%s", "bucket": "%s" }`, target, b.Output)
	}
	// Build the response object.
	var response map[string]interface{}
	json.Unmarshal([]byte(s), &response)
	respondWithJSON(w, http.StatusOK, response)
}

//healthHandler export
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "UP")
}

// RespondWithJSON responds with JSON
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// initialize will initialize the application
func (app *App) initialize() {
	app.initialiseRoutes()
	app.Storage = storage.New()
}

// InitialiseRoutes initialises the routes
func (app *App) initialiseRoutes() {
	app.Router = mux.NewRouter()
	app.Router.Use(loggingMiddleware)
	app.Router.HandleFunc("/", app.methodHandler)
}

// run will run the application
func (app *App) run() {
	log.Println("Starting Server")
	log.Println(http.ListenAndServe(":8080", app.Router))
}

// methodHandler
func (app *App) methodHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		app.labelHandler(w, r)
	case "GET":
		healthHandler(w, r)
	}
}

// main export
func main() {
	app := App{}
	app.initialize()
	app.run()
}

// loggingMiddleware middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)
		log.Printf("%s %s %s", req.Method, req.RequestURI, time.Since(start))
	})
}
