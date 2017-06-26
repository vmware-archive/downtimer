/* Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.

This program and the accompanying materials are made available under
the terms of the under the Apache License, Version 2.0 (the "License”);
you may not use this file except in compliance with the License.

You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

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
