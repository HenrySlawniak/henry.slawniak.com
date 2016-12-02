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
	"bytes"
	"github.com/go-playground/log"
	"net/http"
	"sync"
)

type handler func(http.ResponseWriter, *http.Request) error

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	buf := new(buffer)

	err := h(buf, r)
	if err != nil {
		log.Error(r, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf.Apply(w)
}

type buffer struct {
	bytes.Buffer
	resp    int
	headers http.Header
	once    sync.Once
}

func (b *buffer) Response() int {
	return b.resp
}

func (b *buffer) Header() http.Header {
	b.once.Do(func() {
		b.headers = make(http.Header)
	})
	return b.headers
}

func (b *buffer) WriteHeader(resp int) {
	b.resp = resp
}

func (b *buffer) Apply(w http.ResponseWriter) (n int, err error) {
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
