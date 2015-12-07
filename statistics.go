// Copyright (c) 2015 Henry Slawniak <henry@slawniak.com>
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// Path represents a url path
type Path struct {
	Name  string
	Count int64
}

// Referer represents an http referer
type Referer struct {
	Name  string
	Count int64
}

// RefererList is a slice of Referers
type RefererList []Referer

// PathList is a slice of Paths
type PathList []Path

func (a PathList) Len() int           { return len(a) }
func (a PathList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a PathList) Less(i, j int) bool { return a[i].Count > a[j].Count }

func (a RefererList) Len() int           { return len(a) }
func (a RefererList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RefererList) Less(i, j int) bool { return a[i].Count > a[j].Count }

// Trimmed returns the Name of a Referer trimmed to 64 characters.
func (ref Referer) Trimmed() string {
	runes := bytes.Runes([]byte(ref.Name))
	if len(runes) > 64 {
		return string(runes[:64])
	}
	return string(runes)
}

// SortedPaths returns a list of paths, sorted from highest to lowest value.
func SortedPaths() PathList {
	p := make(PathList, len(Paths))
	i := 0
	for k, v := range Paths {

		p[i] = Path{k, v}
		i++
	}
	sort.Sort(p)
	return p
}

// SortedReferers returns a list of referers, sorted from highest to lowest value.
func SortedReferers() RefererList {
	p := make(RefererList, len(Referers))
	i := 0
	for k, v := range Referers {
		p[i] = Referer{k, v}
		i++
	}
	sort.Sort(p)
	return p
}

// The statistics that we collect
var (
	StartupTime           time.Time
	BytesServed           = int64(0)
	RequestsServed        = int64(0)
	BytesServedSession    = int64(0)
	RequestsServedSession = int64(0)
	Paths                 = make(map[string]int64)
	Referers              = make(map[string]int64)
)

// The saved stats filename
const DataFile = "stats.dat"

// StatsFile represents the saved stats file
type StatsFile struct {
	Paths    map[string]int64
	Referers map[string]int64
	Bytes    int64
	Requests int64
}

// SaveStats saves the stats in memory to file
func SaveStats() error {
	statsfile := StatsFile{}
	if _, err := os.Stat(DataFile); os.IsNotExist(err) {
		os.Create(DataFile)
	}
	file, err := os.Open(DataFile)
	if err != nil {
		return err
	}
	defer file.Close()

	statsfile.Paths = Paths
	statsfile.Referers = Referers
	statsfile.Bytes = BytesServed
	statsfile.Requests = RequestsServed

	out, err := json.MarshalIndent(statsfile, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(DataFile, out, os.ModePerm)
}

// LoadStats loads the stats from file into memory
func LoadStats() error {
	statsfile := StatsFile{}
	if _, err := os.Stat(DataFile); os.IsNotExist(err) {
		os.Create(DataFile)
		return nil
	}
	file, err := os.Open(DataFile)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&statsfile)
	if err != nil {
		return err
	}

	Paths = statsfile.Paths
	Referers = statsfile.Referers
	BytesServed = statsfile.Bytes
	RequestsServed = statsfile.Requests
	return nil
}

// LogRequest increments all appropriate stat counters, then writes to the access log
func LogRequest(req *http.Request, bytes int, response int) {
	ip := strings.Split(req.RemoteAddr, ":")[0]
	if ip == "127.0.0.1" {
		if req.Header.Get("X-Real-IP") != "" {
			ip = req.Header.Get("X-Real-IP")
		}
	}
	AccessLog.Write([]byte(
		fmt.Sprintf("%s [%s] \"%s %s %s %s\" %d %d \"%s\" \"%s\"",
			ip,
			time.Now().Local().Format("02/Jan/2006:15:04:05 -0700"),
			req.Method,
			req.URL.Path,
			req.URL.RawQuery,
			req.Proto,
			response,
			bytes,
			req.Referer(),
			req.UserAgent(),
		) + "\n",
	))
	if !strings.Contains(req.URL.Path, "identicon") && !strings.Contains(req.URL.Path, "api") {
		Paths[req.URL.Path]++
		if req.Referer() != "" {
			Referers[req.Referer()]++
		}
		if req.URL.Query().Get("ref") != "" {
			Referers[req.URL.Query().Get("ref")]++
		}
	}
	BytesServed += int64(bytes)
	RequestsServed++
	BytesServedSession += int64(bytes)
	RequestsServedSession++
	err := SaveStats()
	if err != nil {
		panic(err)
	}
}
