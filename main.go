// main.go
// A web service serving random numbers and strings.

package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/taflaj/util/random"
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds) // I like this particular log format
}

func logIt(r *http.Request) {
	log.Printf("%v %v from %v", r.Method, r.URL.Path, r.RemoteAddr)
}

func helpHandler(w http.ResponseWriter, r *http.Request) {
	logIt(r)
	const page = `
		<!DOCTYPE html>
		<title>%v</title>
		<h1>%v</h1>
		<p><b>Usage:</b> %v/get/<i>type</i>/[<i>length</i>]</p>
		<div>Types:</div>
		<ul>
			<li>number: numerical characters</li>
			<li>hex: hexadecimal integer</li>
			<li>alpha: alphabetical characters</li>
			<li>alphanum: alphabetical and numerical characters</li>
			<li>special: any 7-bit printable characters</li>
			<li>any: any 8-bit characters URL-safe base64-encoded</li>
		</ul>
		<div>Default length is 32</div>
	`
	const title = "Random Number and String Generator"
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, page, title, title, r.Host)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	logIt(r)
	w.Header().Set("Content-Type", "text/plain")
	vars := mux.Vars(r)
	length := 32
	l := vars["length"]
	var err error
	if len(l) > 0 {
		length, err = strconv.Atoi(l)
		if err != nil {
			log.Printf("Error converting %v to int: %#v\n", l, err)
		}
	}
	var result string
	t := vars["type"]
	switch t {
	case "alpha":
		result, err = random.Alpha(length)
	case "alphanum":
		result, err = random.AlphaNum(length)
	case "any":
		result, err = random.Any(length)
	case "hex":
		result, err = random.Hex(length)
	case "number":
		result, err = random.Number(length)
	case "special":
		result, err = random.Special(length)
	}
	if err != nil {
		log.Printf("Error obtaining random %v: %#v\n", t, err)
	}
	fmt.Fprintf(w, result)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", helpHandler)
	r.HandleFunc("/get/{type}/", getHandler)
	r.HandleFunc("/get/{type}/{length:[0-9]+}", getHandler)
	log.Fatal(http.ListenAndServe(":8001", r))
}
