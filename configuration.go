// Copyright (c) 2015 Henry Slawniak <henry@slawniak.com>
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
)

const (
	CONFIG_FILE    = "config.json"
	CONFIG_EXAMPLE = "config.example.json"
)

type Configuration struct {
	Server struct {
		Address string
		Secret  string
		Dburl   string
	}
	Mailgun struct {
		Public      string
		Private     string
		Domain      string
		SenderName  string
		SenderEmail string
	}
	Recaptcha struct {
		Secret  string
		Sitekey string
	}
	Site struct {
		Domain            string
		Title             string
		Description       string
		AllowRegistration bool
	}
}

func (c *Configuration) load() error {
	err := c.ensureConfigExists()
	if err != nil {
		return err
	}

	file, err := os.Open(CONFIG_FILE)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&config)
}

func (c *Configuration) save() error {
	err := c.ensureConfigExists()
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(CONFIG_FILE, out, os.ModePerm)
}

func (c *Configuration) ensureConfigExists() error {
	if _, err := os.Stat(CONFIG_FILE); os.IsNotExist(err) {
		return copyFile(CONFIG_EXAMPLE, CONFIG_FILE)
	} else {
		return nil
	}
}

func copyFile(src string, dest string) error {
	original, err := os.Open(src)
	if err != nil {
		return err
	}
	defer original.Close()

	destination, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, original)
	return err
}
