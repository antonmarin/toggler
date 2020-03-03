package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/schollz/jsonstore"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/add", addHandler)
	r.HandleFunc("/update/{key:[0-9]+}", updateHandler)
	http.Handle("/", r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Printf("Open http://localhost:%s in the browser", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	storage, err := jsonstore.Open("envs.json")
	if err != nil {
		storage = new(jsonstore.JSONStore)
		if err = jsonstore.Save(storage, "envs.json"); err != nil {
			panic(err)
		}
	}

	var state State
	_ = storage.Get("state", &state)

	data := TemplateData{
		State: state,
	}
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		_, _ = w.Write([]byte(fmt.Sprintf("%s", err)))
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		_, _ = w.Write([]byte(fmt.Sprintf("%s", err)))
		return
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	envKey, err := strconv.Atoi(vars["key"])
	if err != nil {
		_, _ = w.Write([]byte(fmt.Sprintf("%s", err)))
		return
	}

	decoder := json.NewDecoder(r.Body)
	var envFromRequest Env
	err = decoder.Decode(&envFromRequest)
	if err != nil {
		panic(err)
	}

	storage, err := jsonstore.Open("envs.json")
	if err != nil {
		panic(err)
	}

	var state State
	_ = storage.Get("state", &state)

	for i, v := range state.Envs {
		if v.Key == envKey {
			_, _ = w.Write([]byte("found"))
			state.Envs[i].Value = envFromRequest.Value
		}
	}

	err = storage.Set("state", state)
	if err != nil {
		_, _ = w.Write([]byte(fmt.Sprintf("%s", err)))
		return
	}

	if err = jsonstore.Save(storage, "envs.json"); err != nil {
		panic(err)
	}

	result, _ := json.Marshal(state.Envs)
	_, _ = w.Write(result)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	storage, err := jsonstore.Open("envs.json")
	if err != nil {
		panic(err)
	}

	var state State
	_ = storage.Get("state", &state)

	var maxKey = 0
	for _, v := range state.Envs {
		if v.Key > maxKey {
			maxKey = v.Key
		}
	}

	maxKey++
	var newEnv = Env{
		Key: maxKey,
		Title: fmt.Sprintf("env%d", maxKey),
		Value: false,
	}
	state.Envs = append(state.Envs, newEnv)

	err = storage.Set("state", state)
	if err != nil {
		_, _ = w.Write([]byte(fmt.Sprintf("%s", err)))
		return
	}

	if err = jsonstore.Save(storage, "envs.json"); err != nil {
		panic(err)
	}
}


type TemplateData struct {
	State State
}

type State struct {
	Envs []Env
}

type Env struct {
	Key   int
	Title string
	Value bool
}