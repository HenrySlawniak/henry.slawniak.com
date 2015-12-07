// Copyright (c) 2015 Henry Slawniak <henry@slawniak.com>
// SPDX-License-Identifier: MIT

package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

type BlogPost struct {
	Id         bson.ObjectId `bson:"_id,omitempty"`
	Title      string
	Slug       string
	Source     template.HTML
	Content    template.HTML
	Date       time.Time
	Author     bson.ObjectId `bson:"_author"`
	Edited     bool
	DateEdited time.Time
	EditedBy   bson.ObjectId `bson:"_editor"`
	Images     []string
	Width      int
}

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func (post BlogPost) Summary() template.HTML {
	bytes := []byte(string(post.Source))
	if len(bytes) < 256 {
		return template.HTML(bluemonday.StrictPolicy().SanitizeBytes(blackfriday.MarkdownCommon(bytes)))
	}
	trimmed := bytes[0:256]
	return template.HTML(bluemonday.StrictPolicy().SanitizeBytes(blackfriday.MarkdownCommon(trimmed)))
}

func (post BlogPost) Store() {
	localsession := session.Copy()
	defer localsession.Close()
	localsession.DB(database).C("blogs").Update(bson.M{"_id": post.Id}, post)
}

func (post BlogPost) CanEdit(user User) bool {
	return (post.Author == user.Id && user.IsBlogAuthor) || user.IsAdmin
}

func (post BlogPost) SlugUrl() string {
	return strings.Join([]string{reverse("blog-read"), post.Slug}, "/")
}

func (post BlogPost) IdUrl() string {
	return strings.Join([]string{reverse("blog-static"), post.Id.Hex()}, "/")
}

func (post BlogPost) GetAuthorAsUser() User {
	localsession := session.Copy()
	defer localsession.Close()
	user := User{}
	localsession.DB(database).C("users").Find(bson.M{"_id": post.Author}).One(&user)
	return user
}

func (post BlogPost) GetEditorAsUser() User {
	localsession := session.Copy()
	defer localsession.Close()
	user := User{}
	localsession.DB(database).C("users").Find(bson.M{"_id": post.EditedBy}).One(&user)
	return user
}

func GetBlogPostWithId(id bson.ObjectId) (BlogPost, error) {
	localsession := session.Copy()
	defer localsession.Close()
	post := BlogPost{}
	err := localsession.DB(database).C("blogs").Find(bson.M{"_id": id}).One(&post)
	if err != nil {
		return post, err
	}
	return post, nil
}

func GetBlogPostWithSlug(slug string) (BlogPost, error) {
	localsession := session.Copy()
	defer localsession.Close()
	post := BlogPost{}
	err := localsession.DB(database).C("blogs").Find(bson.M{"slug": slug}).One(&post)
	if err != nil {
		return post, err
	}
	return post, nil
}

func GetBlogsByAuthor(user *User) (posts []BlogPost, err error) {
	localsession := session.Copy()
	defer localsession.Close()
	err = localsession.DB(database).C("blogs").Find(bson.M{"_author": user.Id}).Sort("-date").All(&posts)
	return
}

func GetBlogPost(slug string) (BlogPost, error) {
	localsession := session.Copy()
	defer localsession.Close()
	post := BlogPost{}
	err := localsession.DB(database).C("blogs").Find(bson.M{"slug": slug}).One(&post)
	if err != nil {
		return post, err
	}
	return post, nil
}

func CountBlogs() int {
	localsession := session.Copy()
	defer localsession.Close()
	count, err := localsession.DB(database).C("blogs").Count()
	if err != nil {
		return 0
	}
	return count
}

func GetBlogs(count int, page int) []BlogPost {
	localsession := session.Copy()
	defer localsession.Close()
	offset := 0
	if page > 1 {
		offset = (page - 1) * count
	}
	blogs := []BlogPost{}
	localsession.DB(database).C("blogs").Find(bson.M{}).Sort("-date").Skip(offset).Limit(count).All(&blogs)
	return blogs
}

func GetBlogsCrono(count int, page int) []BlogPost {
	localsession := session.Copy()
	defer localsession.Close()
	offset := 0
	if page > 1 {
		offset = (page - 1) * count
	}
	blogs := []BlogPost{}
	localsession.DB(database).C("blogs").Find(bson.M{}).Sort("date").Skip(offset).Limit(count).All(&blogs)
	return blogs
}

func CreateBlog(img multipart.File, imageHeader *multipart.FileHeader, w http.ResponseWriter, req *http.Request, ctx *Context) (BlogPost, error) {
	blog := BlogPost{}

	id := bson.NewObjectId()
	title := req.FormValue("title")
	source := req.FormValue("source")
	content := req.FormValue("content")
	date := time.Now().UTC()
	author := ctx.User.Id
	editor := ctx.User.Id
	editdate := time.Now().UTC()

	imgContent, err := ioutil.ReadAll(img)
	if err != nil {
		debug.PrintStack()
		log.Error(err.Error())
		return blog, err
	}
	defer img.Close()

	imgsha1 := fmt.Sprintf("%x", sha1.Sum(imgContent))

	imageFolder := "./static/img/blog/" + id.Hex() + "/"

	os.MkdirAll(imageFolder, 0777)

	err = ioutil.WriteFile(imageFolder+imgsha1+imageHeader.Filename, imgContent, 0664)
	if err != nil {
		debug.PrintStack()
		log.Error(err.Error())
		return blog, err
	}

	diskimg, err := os.Open(imageFolder + imgsha1 + imageHeader.Filename)
	if err != nil {
		debug.PrintStack()
		log.Error(err.Error())
		return blog, err
	}

	srcimage, _, err := image.Decode(diskimg)
	if err != nil {
		debug.PrintStack()
		log.Error(err.Error())
		return blog, err
	}

	WriteJpegImageToFile(imageFolder+imgsha1+".original.100.jpg", 100, srcimage)
	WriteJpegImageToFile(imageFolder+imgsha1+".original.85.jpg", 85, srcimage)
	WriteJpegImageToFile(imageFolder+imgsha1+".original.65.jpg", 65, srcimage)

	images := []string{
		"/assets/img/blog/" + id.Hex() + "/" + imgsha1 + imageHeader.Filename,
		"/assets/img/blog/" + id.Hex() + "/" + imgsha1 + ".original.100.jpg",
		"/assets/img/blog/" + id.Hex() + "/" + imgsha1 + ".original.85.jpg",
		"/assets/img/blog/" + id.Hex() + "/" + imgsha1 + ".original.65.jpg",
	}

	blog.Id = id
	blog.Title = title
	blog.Source = template.HTML(source)
	blog.Content = template.HTML(content)
	blog.Date = date
	blog.Author = author
	blog.EditedBy = editor
	blog.Edited = false
	blog.DateEdited = editdate
	blog.Images = images
	blog.Slug = Slugify(blog.Date.Format("Jan-02-2006-3:04PM") + "-" + blog.Title)
	rand.Seed(time.Now().UnixNano())
	blog.Width = rand.Intn(9) + 1

	localsession := session.Copy()
	defer localsession.Close()

	err = localsession.DB(database).C("blogs").Insert(blog)
	if err != nil {
		debug.PrintStack()
		log.Error(err.Error())
		return blog, err
	}

	return blog, nil
}
