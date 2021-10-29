// app.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"frm/editor/frame"
	"frm/editor/storage"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

// App export
type App struct {
	Router  *mux.Router
	Storage *storage.Storage
}

type Body struct {
	Bucket     string
	Object     string
	Chain      map[string]interface{}
	Output     string
	OutputPath string
}

func (app *App) editorHandler(w http.ResponseWriter, req *http.Request) {
	//READ BODY
	var b Body
	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(req.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//GET IMG
	imgByte, err := app.Storage.Read(b.Bucket, b.Object)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// CREATE PROCESSOR
	chainProcessor := frame.New(b.Chain)

	if chainProcessor == nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// convert []byte to image
	img, _, _ := image.Decode(bytes.NewReader(imgByte))

	//PROCESS CHAIN
	edited := chainProcessor.ProcessChain(img)

	buf := new(bytes.Buffer)

	err = jpeg.Encode(buf, edited, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Resolve name
	filePath := fileNameWithoutExtension(b.Object)
	fileName := filepath.Base(filePath)
	dirName := filepath.Dir(filePath)

	var editName string
	for k, _ := range b.Chain {
		editName = editName + "_" + k
	}
	newName := fmt.Sprintf(`%s/%s/%s_%s.jpg`, dirName, b.OutputPath, fileName, editName)

	obj, err := app.Storage.UploadObject(b.Output, buf.Bytes(), "image/jpeg", newName)
	if err != nil {
		log.Fatal(err)
	}

	s := fmt.Sprintf(`{ "object": "%s", "bucket": "%s" }`, obj, b.Output)

	var response map[string]interface{}
	json.Unmarshal([]byte(s), &response)
	respondWithJSON(w, http.StatusOK, response)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (app *App) initialiseRoutes() {
	app.Router = mux.NewRouter()
	app.Router.HandleFunc("/", app.methodHandler)
}

func (app *App) run() {
	log.Fatal(http.ListenAndServe(":8080", app.Router))
}

func (app *App) methodHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		app.editorHandler(w, r)
	}
}

func main() {
	app := App{}
	app.Storage = storage.New()
	app.initialiseRoutes()
	app.run()
}

func fileNameWithoutExtension(fileName string) string {
	if pos := strings.LastIndexByte(fileName, '.'); pos != -1 {
		return fileName[:pos]
	}
	return fileName
}
