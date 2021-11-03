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
	image := vision.NewImageFromURI(uri)
	annotations, err := client.DetectLabels(ctx, image, nil, 5)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fileName := filepath.Base(b.Object)
	s := fmt.Sprintf(`{ "object": "%s", "bucket": "%s" }`, "No Label", b.Output)
	if len(annotations) > 0 {
		var target string
		for _, annotation := range annotations {
			target = target + "/" + annotation.Description + "/" + fileName
		}
		target = target + "/" + fileName
		app.Storage.CopyFile(b.Bucket, b.Object, b.Output, target)
		s = fmt.Sprintf(`{ "object": "%s", "bucket": "%s" }`, target, b.Output)
	}

	var response map[string]interface{}
	json.Unmarshal([]byte(s), &response)
	respondWithJSON(w, http.StatusOK, response)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "UP")
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (app *App) initialize() {
	app.initialiseRoutes()
	app.Storage = storage.New()
}

func (app *App) initialiseRoutes() {
	app.Router = mux.NewRouter()
	app.Router.Use(loggingMiddleware)
	app.Router.HandleFunc("/", app.methodHandler)
}

func (app *App) run() {
	log.Println("Starting Server")
	log.Println(http.ListenAndServe(":8080", app.Router))
}

func (app *App) methodHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		app.labelHandler(w, r)
	case "GET":
		healthHandler(w, r)
	}
}

func main() {
	app := App{}
	app.initialize()
	app.run()
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)
		log.Printf("%s %s %s", req.Method, req.RequestURI, time.Since(start))
	})
}
