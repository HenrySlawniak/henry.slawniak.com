// Copyright (c) 2016 Henry Slawniak <https://henry.computer/>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"crypto/sha512"
	"flag"
	"fmt"
	"github.com/go-playground/log"
	"github.com/go-playground/log/handlers/console"
	"github.com/gorilla/mux"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	listen = flag.String("listen", "127.0.0.1:2745", "The address to listen on")
	router *mux.Router
)

func init() {
	flag.Parse()
	cLog := console.New()
	cLog.SetTimestampFormat(time.RFC3339)
	log.RegisterHandler(cLog, log.AllLevels...)
}

func main() {
	log.Info("Starting web service")

	setupRouter()

	log.Info("Listening on " + *listen)
	http.ListenAndServe(*listen, router)
}

func setupRouter() {
	log.Info("Setting up router")
	router = mux.NewRouter()
	router.StrictSlash(true)
	router.NotFoundHandler = handler(notFound)
	addRoutes()
}

func addRoutes() {
	log.Info("Addind routes to router")
	router.PathPrefix("/").HandlerFunc(indexHandler())
}

func indexHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if _, err := os.Stat("./client" + path); err == nil {
			serveFile(w, r, "./client"+path)
		} else {
			serveFile(w, r, "./client/index.html")
		}
	})
}

func serveFile(w http.ResponseWriter, r *http.Request, path string) {
	if path == "./client/" {
		path = "./client/index.html"
	}
	content, sum, mod, err := readFile(path)
	if err != nil {
		http.Error(w, "Could not read file", http.StatusInternalServerError)
		fmt.Printf("%s:%s\n", path, err.Error())
		return
	}
	mime := mime.TypeByExtension(filepath.Ext(path))
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "public, no-cache")
	w.Header().Set("Last-Modified", mod.Format(time.RFC1123))
	if r.Header.Get("If-None-Match") == sum {
		w.WriteHeader(http.StatusNotModified)
		w.Header().Set("ETag", sum)
		return
	}
	w.Header().Set("ETag", sum)
	w.Write(content)
}

func readFile(path string) ([]byte, string, time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", time.Now(), err
	}
	defer f.Close()

	stat, err := os.Stat(path)
	if err != nil {
		return nil, "", time.Now(), err
	}

	cont, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, "", time.Now(), err
	}

	return cont, fmt.Sprintf("%x", sha512.Sum512(cont)), stat.ModTime(), nil
}
