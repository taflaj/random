// main.go
// A web service serving random numbers and strings.
// Based on schollz's gist:
// https://gist.github.com/schollz/156d608e8ec26816cedaf06f14d7d692

package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds) // I like this particular log format
	// assert that a cryptographically secure PRNG is available
	buf := make([]byte, 1)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		log.Panicf("Package crypto/rand is unavailable: Read() failed with %#v", err)
	}
	// initialize special character slice
	special = make([]byte, 95)
	for i := 0; i < 95; i++ {
		special[i] = byte(i + 32)
	}
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil { // note that err == nil only if we read len(b) bytes
		return nil, err
	}
	return b, nil
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(domain []byte, length int) (string, error) {
	bytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	if domain == nil { // all bytes: encode it
		return base64.URLEncoding.EncodeToString(bytes), nil
	}
	for i, b := range bytes {
		bytes[i] = domain[b%byte(len(domain))]
	}
	return string(bytes), nil
}

type data struct {
	Title string
	Host  string
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
		<p><b>Usage:</b> %v/<i>command</i>/<i>length</i></p>
		<div>Commands:</div>
		<ul>
			<li>number: numerical characters</li>
			<li>alpha: alphabetical characters</li>
			<li>alphanum: alphabetical and numerical characters</li>
			<li>special: any 7-bit printable characters</li>
			<li>any: any 8-bit characters URL-safe base64-encoded</li>
		</ul>
		<div>Default length is 32</div>
	`
	const title = "Random Number and String Generator"
	fmt.Fprintf(w, page, title, title, r.Host)
}

func generalHandler(w http.ResponseWriter, r *http.Request, chars []byte) {
	logIt(r)
	length := 32
	if p := strings.LastIndex(r.URL.Path, "/"); p != -1 {
		if n, err := strconv.Atoi(r.URL.Path[p+1:]); err == nil {
			if n > 0 {
				length = n
			}
		}
	}
	token, err := GenerateRandomString(chars, length)
	if err != nil {
		log.Panic(err)
	}
	fmt.Fprintf(w, "%v", token)
}

var number = []byte("0123456789")
var alpha = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
var alphanum = append(number, alpha...)
var special []byte

func numberHandler(w http.ResponseWriter, r *http.Request) {
	generalHandler(w, r, number)
}

func alphaHandler(w http.ResponseWriter, r *http.Request) {
	generalHandler(w, r, alpha)
}

func alphanumHandler(w http.ResponseWriter, r *http.Request) {
	generalHandler(w, r, alphanum)
}

func specialHandler(w http.ResponseWriter, r *http.Request) {
	generalHandler(w, r, special)
}

func anyHandler(w http.ResponseWriter, r *http.Request) {
	generalHandler(w, r, nil)
}

func main() {
	http.HandleFunc("/", helpHandler)
	http.HandleFunc("/number/", numberHandler)
	http.HandleFunc("/alpha/", alphaHandler)
	http.HandleFunc("/alphanum/", alphanumHandler)
	http.HandleFunc("/special/", specialHandler)
	http.HandleFunc("/any/", anyHandler)
	log.Fatal(http.ListenAndServe(":8001", nil))
}
