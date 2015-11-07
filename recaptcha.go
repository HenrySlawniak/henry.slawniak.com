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
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const recaptcha_server_name = "https://www.google.com/recaptcha/api/siteverify"

var recaptcha_private_key string

func check(remoteip, response string) (s string) {
	s = ""
	resp, err := http.PostForm(recaptcha_server_name,
		url.Values{"secret": {recaptcha_private_key}, "remoteip": {remoteip}, "response": {response}})
	if err != nil {
		log.Error("Post error: %s", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Read error: could not read body: %s", err)
	} else {
		s = string(body)
	}
	return
}

func RecaptchaConfirm(remoteip, response string) (result bool) {
	result = strings.Contains(check(remoteip, response), "\"success\": true")
	return
}

func RecaptchaInit(key string) {
	recaptcha_private_key = key
}
