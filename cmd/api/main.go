package main

import (
	"net/http"

	"fmt"
	"log"
	"regexp"
)

var emailRegex = regexp.MustCompile(
	`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`,
)

func get_by_category(w http.ResponseWriter, r *http.Request) {
	category := r.PathValue("category")

	fmt.Fprintf(w, "Query: category %s", category)
}

func get_by_maintainer(w http.ResponseWriter, r *http.Request) {
	maintainer := r.PathValue("maintainer")

	if !emailRegex.MatchString(maintainer) {
		http.Error(w, "missing or invalid parameter: maintainer", http.StatusBadRequest)
	}

	fmt.Fprintf(w, "Query: maintainer %s", maintainer)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /updates/category/{category}", get_by_category)
	mux.HandleFunc("GET /updates/maintainer/{maintainer}", get_by_maintainer)

	err := http.ListenAndServe(":4000", mux)

	if err != nil {
		log.Fatal(err)
	}
}
