// Copyright (c) 2015 Henry Slawniak <henry@slawniak.com>
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"net/http"
	"sync"
)

type handler func(http.ResponseWriter, *http.Request, *Context, bool) error

func (h handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//create the context
	ctx, err := NewContext(req)
	if err != nil {
		log.Info("Error creating context!")
		log.Info(err.Error())
		panic(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer ctx.Close()

	//run the handler and grab the error, and report it
	buf := new(Buffer)
	pjax := req.Header.Get("X-PJAX") != ""
	err = h(buf, req, ctx, pjax)
	if err != nil {
		log.Info("Error creating buffer!")
		log.Info(err.Error())
		panic(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//save the session
	if err = ctx.Session.Save(req, buf); err != nil {
		log.Info("Error saving session!")
		log.Info(err.Error())
		panic(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, _ := buf.Apply(w)
	go LogRequest(req, bytes, buf.Response())
}

type Buffer struct {
	bytes.Buffer
	resp    int
	headers http.Header
	once    sync.Once
}

func (b *Buffer) Response() int {
	return b.resp
}

func (b *Buffer) Header() http.Header {
	b.once.Do(func() {
		b.headers = make(http.Header)
	})
	return b.headers
}

func (b *Buffer) WriteHeader(resp int) {
	b.resp = resp
}

func (b *Buffer) Apply(w http.ResponseWriter) (n int, err error) {
	if len(b.headers) > 0 {
		h := w.Header()
		for key, val := range b.headers {
			h[key] = val
		}
	}
	if b.resp > 0 {
		w.WriteHeader(b.resp)
	} else {
		b.resp = 200
	}
	n, err = w.Write(b.Bytes())
	return
}
