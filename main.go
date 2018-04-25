package main

import (
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates",
			t.filename)))
	})
	t.templ.Execute(w, r)
}

func main() {
	var addr string = ":8080"
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/home", &templateHandler{filename: "index.html"})
	http.HandleFunc("/createroom", createRoom)
	http.HandleFunc("/joinroom/", joinRoom)
	http.Handle("/room/", &templateHandler{filename: "room.html"})
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
