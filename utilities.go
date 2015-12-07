// Copyright (c) 2015 Henry Slawniak <henry@slawniak.com>
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"golang.org/x/text/unicode/norm"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"unicode"
)

var lat = []*unicode.RangeTable{unicode.Letter, unicode.Number}

func reverse(name string, things ...interface{}) string {
	//convert the things to strings
	strs := make([]string, len(things))
	for i, th := range things {
		strs[i] = fmt.Sprint(th)
	}
	//grab the route
	u, err := router.GetRoute(name).URL(strs...)
	if err != nil {
		panic(err)
	}
	return u.Path
}

func Slugify(s string) string {
	buf := make([]rune, 0, len(s))
	dash := false
	for _, r := range norm.NFKD.String(s) {
		switch {
		case unicode.IsOneOf(lat, r):
			buf = append(buf, unicode.ToLower(r))
			dash = true
		case dash:
			buf = append(buf, '-')
			dash = false
		}
	}
	if i := len(buf) - 1; i >= 0 && buf[i] == '-' {
		buf = buf[:i]
	}
	return string(buf)
}

func WriteJpegImageToFile(path string, quality int, img image.Image) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModeDir)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, &jpeg.Options{Quality: quality})
}
