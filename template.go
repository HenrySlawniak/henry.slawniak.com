// Copyright (c) 2015 Henry Slawniak <henry@slawniak.com>
// SPDX-License-Identifier: MIT

package main

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/dustin/go-humanize"
	"html/template"
	"image/color"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var cachedTemplates = map[string]*template.Template{}
var pjaxTemplates = map[string]*template.Template{}
var cachedMutex sync.Mutex

var funcs = template.FuncMap{
	"reverse":           reverse,
	"timestamptostring": timestamptostring,
	"fdate":             fdate,
	"ftimeago":          ftimeago,
	"f8601":             f8601,
	"formatBytes":       formatBytes,
	"formatColor":       formatColor,
	"possessive":        possessive,
	"flargenum":         flargenum,
	"join":              join,
	"json":              indentjson,
}

func indentjson(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "  ")
	return string(s)
}

func join(s []string) string {
	return strings.Join(s, ",")
}

func flargenum(n int64) string {
	if n > 1000000000 {
		return fmt.Sprintf("%.2f", float64(n)/float64(1000000000)) + "B"
	} else if n > 1000000 {
		return fmt.Sprintf("%.2f", float64(n)/float64(1000000)) + "M"
	} else if n > 1000 {
		return fmt.Sprintf("%.2f", float64(n)/float64(1000)) + "K"
	} else {
		return fmt.Sprintf("%d", n)
	}
}

func possessive(s string) string {
	if s[len(s)-1] == 's' {
		return s + "'"
	}
	return s + "'s"
}

func formatColor(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%X%X%X", uint8(r), uint8(g), uint8(b))
}

func formatBytes(i int64) string {
	return humanize.Bytes(uint64(i))
}

func f8601(t time.Time) string {
	return t.Format(time.RFC3339)
}

func fdate(t time.Time) template.HTML {
	format := "<time is=\"local-time\" datetime=\"%s\">%s</time>"
	return template.HTML(fmt.Sprintf(format, t.UTC().Format(time.RFC3339), t.Format("January 2, 2006 at 3:04 PM")))
}

func ftimeago(t time.Time) template.HTML {
	format := "<time is=\"relative-time\" datetime=\"%s\">%s</time>"
	return template.HTML(fmt.Sprintf(format, t.UTC().Format(time.RFC3339), t.UTC().Format("Mon Jan 2")))
}

func timestamptostring(stamp int64) string {
	return fmt.Sprintln(time.Unix(stamp, 0))
}

func md5Data(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}

func sha1Data(data []byte) string {
	return fmt.Sprintf("%x", sha1.Sum(data))
}

func T(name string, pjax bool) *template.Template {
	cachedMutex.Lock()
	defer cachedMutex.Unlock()

	if pjax {
		if t, ok := pjaxTemplates[name]; ok {
			return t
		}
		t := template.New("_base.pjax.html").Funcs(funcs)

		t = template.Must(t.ParseFiles(
			"templates/_base.pjax.html",
			filepath.Join("templates", name),
		))
		pjaxTemplates[name] = t

		return t
	}

	if t, ok := cachedTemplates[name]; ok {
		return t
	}

	t := template.New("_base.html").Funcs(funcs)

	t = template.Must(t.ParseFiles(
		"templates/_nav.html",
		"templates/_base.html",
		filepath.Join("templates", name),
	))
	// cachedTemplates[name] = t

	return t
}
