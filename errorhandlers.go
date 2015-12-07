// Copyright (c) 2015 Henry Slawniak <henry@slawniak.com>
// SPDX-License-Identifier: MIT

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
