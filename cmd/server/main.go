package main

import (
	"embed"
	"net/http"

	"github.com/gorilla/mux"
)

//go:embed public/*
var public embed.FS

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.PathPrefix("/public").Handler(http.FileServer(http.FS(public)))

	http.ListenAndServe(":3000", router)
}
