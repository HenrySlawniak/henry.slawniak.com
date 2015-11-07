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
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/mailgun/mailgun-go"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"net/url"
	"strings"
	"time"
)

type User struct {
	Id              bson.ObjectId `bson:"_id,omitempty"`
	Email           string
	Username        string
	Password        []byte
	EmailVerified   bool
	EmailVerifyHash string
	DisplayName     string
	IsAdmin         bool
	IsBlogAuthor    bool
	LastVisit       time.Time
	Balance         int64
	UsingGravatar   bool
	Bio             struct {
		Source  template.HTML
		Content template.HTML
	}
	Awards []struct {
		Name    string
		Icon    string
		Awarded time.Time
		Color   string
	}
}

func (u User) Avatar() string {
	if u.UsingGravatar {
		return u.Gravatar()
	}
	return u.Identicon()
}

func (u User) Gravatar() string {
	return "https://www.gravatar.com/avatar/" + fmt.Sprintf("%x", md5.Sum([]byte(strings.Trim(strings.ToLower(u.Email), " ")))) + "?s=420&d=" + url.QueryEscape("https://fortkickass.co"+u.Identicon())
}

func (u User) Identicon() string {
	return reverse("identicon") + "/" + u.GetInfoHash()
}

func (u User) GetInfoHash() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%x%x%x", md5.Sum([]byte(u.Email)), md5.Sum([]byte(u.Username)), md5.Sum([]byte(u.Id.Hex()))))))
}

func (u User) SendVerificationEmail() error {
	localsession := session.Copy()
	defer localsession.Close()
	if u.EmailVerified {
		return errors.New("Already Verified")
	}
	if u.EmailVerifyHash == "" {
		u.EmailVerifyHash = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%x%s", u.GetInfoHash(), time.Now().UTC()))))
		localsession.DB(database).C("users").Update(bson.M{"_id": u.Id}, u)
	}
	gun := mailgun.NewMailgun(config.Mailgun.Domain, config.Mailgun.Private, config.Mailgun.Public)
	m := mailgun.NewMessage(
		fmt.Sprintf("%s <%s>", config.Mailgun.SenderName, config.Mailgun.SenderEmail),
		fmt.Sprintf("Verify Your Email With %s", config.Site.Title),
		fmt.Sprintf("Click below to verify this email address for %s on %s\nhttp://%s/account/verifyemail/%s", u.Username, config.Site.Title, config.Site.Domain, u.EmailVerifyHash),
		fmt.Sprintf("%s <%s>", u.Username, u.Email),
	)
	_, _, err := gun.Send(m)
	return err
}

func (u *User) Store() {
	localsession := session.Copy()
	defer localsession.Close()
	localsession.DB(database).C("users").Update(bson.M{"_id": u.Id}, u)
}

func (u *User) SetPassword(password string) {
	hpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err) //this is a panic because bcrypt errors on invalid costs
	}
	u.Password = hpass
}

func (u *User) SetEmailVerified() {
	localsession := session.Copy()
	defer localsession.Close()
	u.EmailVerified = true
	localsession.DB(database).C("users").Update(bson.M{"_id": u.Id}, u)
}

func (u *User) UpdateLastVisit() time.Time {
	u.LastVisit = time.Now().UTC()
	u.Store()
	return u.LastVisit
}

func GetUserByName(name string) (u *User, err error) {
	localsession := session.Copy()
	defer localsession.Close()
	err = localsession.DB(database).C("users").Find(bson.M{"username": name}).One(&u)
	if err != nil {
		return
	}
	return
}

func GetAllUsers() []User {
	localsession := session.Copy()
	defer localsession.Close()
	users := []User{}
	localsession.DB(database).C("users").Find(bson.M{}).Sort("_id").All(&users)
	return users
}

func Login(username, password string) (u *User, err error) {
	localsession := session.Copy()
	defer localsession.Close()
	err = localsession.DB(database).C("users").Find(bson.M{"username": username}).One(&u)
	if err != nil {
		return
	}

	err = bcrypt.CompareHashAndPassword(u.Password, []byte(password))
	if err != nil {
		u = nil
	}
	return
}
