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
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strings"
	"unicode/utf8"
)

func LoginFormHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) error {
	return T("pages/login/login.html", pjax).Execute(w, map[string]interface{}{
		"ctx": ctx,
	})
}

func RegisterFormHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) error {
	return T("pages/login/register.html", pjax).Execute(w, map[string]interface{}{
		"ctx":     ctx,
		"sitekey": config.Recaptcha.Sitekey,
	})
}

func LogoutHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) error {
	delete(ctx.Session.Values, "user")
	http.Redirect(w, req, reverse("index"), http.StatusSeeOther)
	return nil
}

func LoginHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) error {
	username, password := req.FormValue("username"), req.FormValue("password")

	user, err := Login(username, password)
	if err != nil {
		ctx.Session.AddFlash("Invalid Username/Password")
		return LoginFormHandler(w, req, ctx, pjax)
	}

	ctx.Session.Values["user"] = user.Id
	ctx.Session.Values["name"] = user.Username
	http.Redirect(w, req, reverse("index"), http.StatusSeeOther)
	return nil
}

func RegisterHandler(w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) error {
	username, password, email := req.FormValue("username"), req.FormValue("password"), req.FormValue("email")

	if utf8.RuneCountInString(username) > 16 {
		ctx.Session.AddFlash("Username too long.")
		return RegisterFormHandler(w, req, ctx, pjax)
	}
	if utf8.RuneCountInString(username) < 3 {
		ctx.Session.AddFlash("Username too short.")
		return RegisterFormHandler(w, req, ctx, pjax)
	}
	if !REGEX_NAME.MatchString(username) {
		ctx.Session.AddFlash("Invalid username.")
		return RegisterFormHandler(w, req, ctx, pjax)
	}

	if utf8.RuneCountInString(password) > 128 {
		ctx.Session.AddFlash("Password too long.")
		return RegisterFormHandler(w, req, ctx, pjax)
	}
	if utf8.RuneCountInString(password) < 1 {
		ctx.Session.AddFlash("Password too short.")
		return RegisterFormHandler(w, req, ctx, pjax)
	}

	if utf8.RuneCountInString(email) < 5 {
		ctx.Session.AddFlash("Email too short.")
		return RegisterFormHandler(w, req, ctx, pjax)
	}
	if !REGEX_EMAIL.MatchString(email) {
		ctx.Session.AddFlash("Invalid email.")
		return RegisterFormHandler(w, req, ctx, pjax)
	}

	ip := strings.Split(req.RemoteAddr, ":")[0]
	if ip == "127.0.0.1" {
		if req.Header.Get("X-Real-IP") != "" {
			ip = req.Header.Get("X-Real-IP")
		}
	}
	result := false
	response := req.PostFormValue("g-recaptcha-response")
	if response != "" {
		result = RecaptchaConfirm(ip, response)
	}

	if !result {
		ctx.Session.AddFlash("Recaptcha failed.")
		return RegisterFormHandler(w, req, ctx, pjax)
	}

	u := &User{
		Username:    username,
		DisplayName: username,
		Id:          bson.NewObjectId(),
		Email:       email,
	}
	u.SetPassword(password)

	localsession := session.Copy()
	defer localsession.Close()

	if err := localsession.DB(database).C("users").Insert(u); err != nil {
		ctx.Session.AddFlash("Problem registering user.")
		return RegisterFormHandler(w, req, ctx, pjax)
	}

	ctx.Session.Values["user"] = u.Id
	ctx.Session.Values["name"] = u.Username
	u.SendVerificationEmail()
	http.Redirect(w, req, reverse("index"), http.StatusSeeOther)
	return nil
}
