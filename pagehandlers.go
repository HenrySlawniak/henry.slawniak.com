// Copyright (c) 2015 Henry Slawniak <henry@slawniak.com>
// SPDX-License-Identifier: MIT

package main

import (
	"net/http"
)

func IndexPageHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) (err error) {
	return T("pages/index.html", pjax).Execute(w, map[string]interface{}{
		"ctx":   ctx,
		"blogs": GetBlogs(6, 0),
	})
}

func BioPageHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) (err error) {
	return T("pages/bio.html", pjax).Execute(w, map[string]interface{}{
		"ctx": ctx,
	})
}

func ClockPageHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) (err error) {
	return T("pages/clock.html", pjax).Execute(w, map[string]interface{}{
		"ctx": ctx,
	})
}
