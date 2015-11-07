// Copyright (c) 2015 Henry Slawniak <http://fortkickass.co/>
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
	"net/http"
)

func SetupErrorMessages() {
	ErrorMessages = map[int]string{
		http.StatusNotFound:            "That page doesn't exist",
		http.StatusUnauthorized:        "You're not Dave! You need to Login!",
		http.StatusInternalServerError: "Something is going horribly wrong!",
		http.StatusBadRequest:          "You can't do that!",
	}
}

func httpError(code int, w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) error {
	w.WriteHeader(code)
	message := ""
	if _, ok := ErrorMessages[code]; ok {
		message = ErrorMessages[code]
	}
	return T("_error.html", pjax).Execute(w, map[string]interface{}{
		"ctx":     ctx,
		"code":    code,
		"message": message,
	})
}

func NotFoundHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) error {
	return httpError(http.StatusNotFound, w, req, ctx, pjax)
}

func NotAuthedHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) error {
	return httpError(http.StatusUnauthorized, w, req, ctx, pjax)
}

func InternalErrorHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) error {
	return httpError(http.StatusInternalServerError, w, req, ctx, pjax)
}

func BadRequestHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) error {
	return httpError(http.StatusBadRequest, w, req, ctx, pjax)
}
