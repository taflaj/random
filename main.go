// main.go
// A web service serving random numbers and strings.
// Based on schollz's gist:
// https://gist.github.com/schollz/156d608e8ec26816cedaf06f14d7d692

package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
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
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		log.Panicf("crypto/rand is unavailable: Read() failed with %#v", err)
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
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b) // note that err == nil only if we read len(b) bytes
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(chars []byte, length int) (string, error) {
	bytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	if chars == nil { // all bytes: encode it
		return base64.URLEncoding.EncodeToString(bytes), nil
	}
	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
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
	t, err := template.ParseFiles("html/help.html")
	if err != nil {
		log.Panic(err)
	}
	if err = t.Execute(w, &data{Title: "Random Number and String Generator", Host: r.Host}); err != nil {
		log.Panic(err)
	}
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
