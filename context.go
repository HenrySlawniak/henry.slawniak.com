// Copyright (c) 2015 Henry Slawniak <henry@slawniak.com>
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/gorilla/sessions"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

type Context struct {
	Session *sessions.Session
	User    *User
	Site    interface{}
}

func (c *Context) Close() {
	// c.Database.Session.Close()
}

func NewContext(req *http.Request) (*Context, error) {
	sess, err := store.Get(req, "session")
	ctx := &Context{
		Session: sess,
		Site:    config.Site,
	}
	if err != nil {
		return ctx, err
	}

	//try to fill in the user from the session
	if uid, ok := sess.Values["user"].(bson.ObjectId); ok {
		localsession := session.Copy()
		defer localsession.Close()
		err = localsession.DB(database).C("users").Find(bson.M{"_id": uid}).One(&ctx.User)
	}

	return ctx, err
}
