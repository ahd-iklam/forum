package main

import (
	"fmt"
	"net/http"
)

func filterPosts(w http.ResponseWriter, r *http.Request) {
	categories := r.Form["category"]
	appt := r.FormValue("appt")
	likes := r.FormValue("likes")

	fmt.Println(categories)
	fmt.Println( appt)
	fmt.Println(likes)
}
