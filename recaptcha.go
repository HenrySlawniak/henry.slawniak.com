// Copyright (c) 2015 Henry Slawniak <henry@slawniak.com>
// SPDX-License-Identifier: MIT

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
