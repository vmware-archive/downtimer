package main

import (
	"fmt"
	"net/http"
)

func main() {

	fs := http.FileServer(http.Dir("public"))

	http.Handle("/", fs)

	fmt.Println("Starting Server on 3000")

	http.ListenAndServe(":3000", nil)
}
