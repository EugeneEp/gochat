package main

import (
	"flag"
	"log"
	"net/http"
	"encoding/json"
	"os"
	"io"
	"gochat/chat"
	"github.com/gorilla/mux"
	"html/template"
	"github.com/rs/cors"
	"strings"
)

var addr = flag.String("addr", ":8080", "http service address")

func outputHTML(w http.ResponseWriter, filename string, data interface{}) {
    t, err := template.ParseFiles(filename)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    if err := t.Execute(w, data); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
}

func serveChat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if !strings.Contains(r.URL.Path, "/chat") {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	arr := map[string]interface{}{"ChatName": vars["chat_name"]}
    outputHTML(w, "chat.html", arr)
}

func serveHome(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func main() {
	flag.Parse()
	hub := chat.NewHub()
	go hub.Run()
	log.Println("test")
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:8080"},
		AllowedMethods: []string{"GET", "HEAD", "POST", "PUT", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})
	r := mux.NewRouter()
	r.HandleFunc("/", serveHome).Methods("GET")
	r.PathPrefix("/files/").Handler(http.StripPrefix("/files/", http.FileServer(http.Dir("files/"))))
	r.HandleFunc("/chat/{chat_name}", serveChat).Methods("GET")
	r.HandleFunc("/ws/{channel_name}", func(w http.ResponseWriter, r *http.Request) {
		chat.ServeWs(hub, w, r)
	})

	r.HandleFunc("/file/{chat_name}", func(w http.ResponseWriter, r *http.Request) {
		loadFile(w, r)
		response := map[string]interface{}{"success":true}
		js, _ := json.Marshal(response)
		w.Header().Set("Content-Type","application/json")
		w.Write(js)
	}).Methods("POST")

	err := http.ListenAndServe(*addr, c.Handler(r))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func loadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	file, header, err := r.FormFile("file")
    fileName := "files/" + vars["chat_name"] + "/" + header.Filename
    if err != nil {
        log.Println(err)
    }
    defer file.Close()

    log.Println("test2")

    f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        log.Println(err)
    }
    defer f.Close()
    _, _ = io.Copy(f, file)
}

/*

file, _, err := r.FormFile("file")
    fileName := "files/" + vars["chat_name"] + "/" + data["filaname"] + ".png"
    if err != nil {
        log.Println(err)
    }
    defer file.Close()

    log.Println("test2")

    f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        log.Println(err)
    }
    defer f.Close()
    _, _ = io.Copy(f, file)

    response := map[string]interface{}{"success":true}
	js, _ := json.Marshal(response)
	w.Header().Set("Content-Type","application/json")
	w.Write(js)

*/