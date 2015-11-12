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
	"crypto/sha1"
	// "errors"
	"fmt"
	// "github.com/muesli/smartcrop"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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
}

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
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

func CreateBlog(img multipart.File, imageHeader *multipart.FileHeader, w http.ResponseWriter, req *http.Request, ctx *Context, pjax bool) (BlogPost, error) {
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

	os.MkdirAll(imageFolder, os.ModeDir)

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

	// analyzer := smartcrop.NewAnalyzer()
	// wideCrop, err := analyzer.FindBestCrop(srcimage, 1568, 588)
	// if err != nil {
	// 	return blog, err
	// }
	//
	// posterCrop, err := analyzer.FindBestCrop(srcimage, 478, 388)
	// if err != nil {
	// 	return blog, err
	// }
	//
	// sub, ok := img.(SubImager)
	// if ok {
	// 	wide := sub.SubImage(image.Rect(wideCrop.X, wideCrop.Y, wideCrop.Width+wideCrop.X, wideCrop.Height+wideCrop.Y))
	// 	poster := sub.SubImage(image.Rect(posterCrop.X, posterCrop.Y, posterCrop.Width+posterCrop.X, posterCrop.Height+posterCrop.Y))
	//
	// 	WriteJpegImageToFile(imageFolder+imgsha1+".1568x588.100.jpg", 100, wide)
	// 	WriteJpegImageToFile(imageFolder+imgsha1+".1568x588.85.jpg", 85, wide)
	// 	WriteJpegImageToFile(imageFolder+imgsha1+".1568x588.65.jpg", 65, wide)
	//
	// 	WriteJpegImageToFile(imageFolder+imgsha1+".478x388.100.jpg", 100, poster)
	// 	WriteJpegImageToFile(imageFolder+imgsha1+".478x388.85.jpg", 85, poster)
	// 	WriteJpegImageToFile(imageFolder+imgsha1+".478x388.65.jpg", 65, poster)
	// } else {
	// 	return blog, errors.New("No Subimage support")
	// }

	images := []string{
		"/assets/img/blog/" + id.Hex() + "/" + imgsha1 + ".original.100.jpg",
		"/assets/img/blog/" + id.Hex() + "/" + imgsha1 + ".original.85.jpg",
		"/assets/img/blog/" + id.Hex() + "/" + imgsha1 + ".original.65.jpg",
		"/assets/img/blog/" + id.Hex() + "/" + imgsha1 + ".1568x588.100.jpg",
		"/assets/img/blog/" + id.Hex() + "/" + imgsha1 + ".1568x588.85.jpg",
		"/assets/img/blog/" + id.Hex() + "/" + imgsha1 + ".1568x588.65.jpg",
		"/assets/img/blog/" + id.Hex() + "/" + imgsha1 + ".478x388.100.jpg",
		"/assets/img/blog/" + id.Hex() + "/" + imgsha1 + ".478x388.85.jpg",
		"/assets/img/blog/" + id.Hex() + "/" + imgsha1 + ".478x388.65.jpg",
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
	blog.Slug = Slugify(blog.Title)

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

func WriteJpegImageToFile(path string, quality int, img image.Image) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModeDir)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		debug.PrintStack()
		log.Error(err.Error())
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, &jpeg.Options{Quality: quality})
}
