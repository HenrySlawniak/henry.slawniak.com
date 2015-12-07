// Copyright (c) 2015 Henry Slawniak <henry@slawniak.com>
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/gob"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/op/go-logging"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/natefinch/lumberjack.v2"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"time"
)

const (
	VERSION = "0.0.1"
)

var (
	config        = &Configuration{}
	session       *mgo.Session
	database      string
	router        *mux.Router
	store         sessions.Store
	REGEX_EMAIL   *regexp.Regexp
	REGEX_NAME    *regexp.Regexp
	ErrorMessages map[int]string
	AccessLog     *lumberjack.Logger
)

var log = logging.MustGetLogger("fortkickass.co")
var format = "[%{time:15:04:05.0000}] %{level:.4s} %{message}"

func setupConfig() {
	err := config.load()
	if err != nil {
		fmt.Printf("Error loading config: %s\n", err)
		return
	}
}

func setupLog(logBackend *logging.LogBackend) {
	logging.SetBackend(logBackend)
	logging.SetFormatter(logging.MustStringFormatter(format))
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	setupLog(logBackend)
	setupConfig()
	SetupErrorMessages()
	gob.Register(bson.ObjectId(""))
	RecaptchaInit(config.Recaptcha.Secret)

	var err error
	session, err = mgo.Dial(config.Server.Dburl)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	database = session.DB("").Name

	if err := session.DB("").C("users").EnsureIndex(mgo.Index{
		Key:    []string{"username"},
		Unique: true,
	}); err != nil {
		log.Fatal(err)
	}

	REGEX_EMAIL, err = regexp.Compile(`^[_a-z0-9-]+(\.[_a-z0-9-]+)*@[a-z0-9-]+(\.[a-z0-9-]+)*(\.[a-z]{2,3})$`)
	if err != nil {
		log.Fatal(err)
	}

	REGEX_NAME, err = regexp.Compile(`[a-zA-Z0-9_]+`)
	if err != nil {
		log.Fatal(err)
	}

	store = sessions.NewCookieStore([]byte(config.Server.Secret))

	router = mux.NewRouter()
	router.StrictSlash(true)
	router.NotFoundHandler = handler(NotFoundHandler)

	router.Path("/").Handler(handler(IndexPageHandler)).Name("index").Methods("GET")
	router.Path("/bio").Handler(handler(BioPageHandler)).Name("bio").Methods("GET")
	router.Path("/blog").Handler(handler(BlogIndexHandler)).Name("blog").Methods("GET")

	router.Path("/blog/write").Handler(handler(BlogWriteFormHandler)).Name("blog-write").Methods("GET")
	router.Path("/blog/write").Handler(handler(BlogWriteHandler)).Methods("POST")

	router.Path("/blog/edit").Name("blog-edit")
	// router.Path("/blog/edit/{id}").Handler(handler(BlogEditFormHandler)).Methods("GET")
	// router.Path("/blog/edit").Handler(handler(BlogEditHandler)).Methods("POST")

	router.Path("/blog/read").Name("blog-read")
	router.Path("/blog/read/{slug}").Handler(handler(BlogReadHandler)).Methods("GET")

	router.Path("/blog/static").Name("blog-static")
	router.Path("/blog/static/{id}").Handler(handler(BlogStaticHandler)).Methods("GET")

	router.Path("/login").Handler(handler(LoginHandler)).Name("login").Methods("POST")
	router.Path("/login").Handler(handler(LoginFormHandler)).Methods("GET")
	router.Path("/register").Handler(handler(RegisterHandler)).Name("register").Methods("POST")
	router.Path("/register").Handler(handler(RegisterFormHandler)).Methods("GET")
	router.Path("/logout").Handler(handler(LogoutHandler)).Name("logout").Methods("GET")

	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./static/"))))

	os.MkdirAll("./static/img/blog/", os.ModeDir)

	log.Info(fmt.Sprintf("Listening on %s", config.Server.Address))
	AccessLog = &lumberjack.Logger{
		Filename:   "logs/access.log",
		MaxSize:    500, // megabytes
		MaxBackups: 0,
	}
	defer AccessLog.Close()

	err = LoadStats()
	if err != nil {
		panic(err)
	}

	StartupTime = time.Now()
	http.Handle("/", router)
	err = http.ListenAndServe(config.Server.Address, nil)
	log.Critical(err.Error())
}
