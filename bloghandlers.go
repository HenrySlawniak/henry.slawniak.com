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
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

func BlogIndexHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) (err error) {
	return nil
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

	blog, err := CreateBlog(file, header, w, req, ctx, pjax)
	if err != nil {
		ctx.Session.AddFlash(err.Error())
		return BlogWriteFormHandler(w, req, ctx, pjax)
	}

	http.Redirect(w, req, blog.IdUrl(), http.StatusFound)
	return nil
}
