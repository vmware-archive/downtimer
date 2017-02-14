package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

func handler(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("public/index.html") // Parse template file.
	if err != nil {
		fmt.Println(err)
	}

	files, _ := filepath.Glob("public/*.csv")

	t.Execute(w, files)

}

func main() {

	fs := http.FileServer(http.Dir("public"))

	http.Handle("/public/", http.StripPrefix("/public/", fs))
	http.HandleFunc("/index.html", handler)

	fmt.Println("Starting Server on 3000")

	http.ListenAndServe(":3000", nil)
}
