// Copyright (c) 2015 Henry Slawniak <henry@slawniak.com>
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

func BlogIndexHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) (err error) {
	return T("pages/blog/index.html", pjax).Execute(w, map[string]interface{}{
		"ctx":   ctx,
		"blogs": GetBlogs(50, 0),
	})
}

func BlogWriteFormHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) (err error) {
	if ctx.User == nil || (!ctx.User.IsBlogAuthor && !ctx.User.IsAdmin) {
		return NotAuthedHandler(w, req, ctx, pjax)
	}
	return T("pages/blog/write.html", pjax).Execute(w, map[string]interface{}{
		"ctx": ctx,
	})
}

func BlogReadHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) (err error) {
	vars := mux.Vars(req)
	slug := vars["slug"]

	post, err := GetBlogPostWithSlug(slug)
	if err != nil {
		return NotFoundHandler(w, req, ctx, pjax)
	}

	return T("pages/blog/read.html", pjax).Execute(w, map[string]interface{}{
		"ctx":  ctx,
		"post": post,
	})
}

func BlogStaticHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) (err error) {
	vars := mux.Vars(req)
	id := vars["id"]

	if bson.IsObjectIdHex(id) {
		realid := bson.ObjectIdHex(id)
		post, err := GetBlogPostWithId(realid)
		if err != nil {
			return NotFoundHandler(w, req, ctx, pjax)
		}

		http.Redirect(w, req, post.SlugUrl(), http.StatusFound)
		return nil
	}

	return BadRequestHandler(w, req, ctx, pjax)
}

func BlogWriteHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) (err error) {
	if ctx.User == nil || (!ctx.User.IsBlogAuthor && !ctx.User.IsAdmin) {
		return NotAuthedHandler(w, req, ctx, pjax)
	}

	file, header, err := req.FormFile("blog-image")
	if err != nil {
		ctx.Session.AddFlash(err.Error())
		return BlogWriteFormHandler(w, req, ctx, pjax)
	}

	blog, err := CreateBlog(file, header, w, req, ctx)
	if err != nil {
		ctx.Session.AddFlash(err.Error())
		return BlogWriteFormHandler(w, req, ctx, pjax)
	}

	http.Redirect(w, req, blog.IdUrl(), http.StatusFound)
	return nil
}
