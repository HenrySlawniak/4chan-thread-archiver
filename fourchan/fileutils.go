// Package fourchan provides an API client for 4chan.org

// Copyright (c) 2015 Henry Slawniak <henry@slawniak.com>
// SPDX-License-Identifier: MIT

package fourchan

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.22 Safari/537.36"

// DownloadFile will write the file to disc from url
// It will also check mime types unless '*' is provided for mime
func DownloadFile(fileloc, url, mime, useragent string) (int64, error) {
	os.MkdirAll(filepath.Dir(fileloc), os.ModeDir)
	if useragent == "" {
		useragent = USER_AGENT
	}
	client := &http.Client{}
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", useragent)
	req.Header.Set("Accepts", mime)

	head, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	if head.StatusCode != 200 {
		return 0, errors.New(fmt.Sprintf("Bad response, expected 200 got %d (%s)", head.StatusCode, url))
	}
	if mime != "*" && head.Header.Get("Content-Type") != mime {
		return 0, errors.New(fmt.Sprintf("Wrong MIME, expected %s got %s (%s)", mime, head.Header.Get("Content-Type"), url))
	}
	if !FileExists(fileloc) {
		return writeUrlToFile(fileloc, url, mime, useragent)
	} else {
		contlength, err := strconv.ParseInt(head.Header.Get("Content-Length"), 10, 0)
		if err != nil {
			return 0, err
		}
		file, err := os.Open(fileloc)
		if err != nil {
			return 0, err
		}
		defer file.Close()

		fstat, err := file.Stat()
		if err != nil {
			return 0, err
		}

		if fstat.Size() == contlength {
			fmt.Printf("Skipping %s\n", fileloc)
			return 0, nil
		} else {
			return writeUrlToFile(fileloc, url, mime, useragent)
		}
	}
}

func writeUrlToFile(fileloc, url, mime, useragent string) (int64, error) {
	fmt.Printf("Writing %s\n", fileloc)
	var out *os.File

	if useragent == "" {
		useragent = USER_AGENT
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", useragent)
	req.Header.Set("Accepts", mime)

	get, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer get.Body.Close()

	if FileExists(fileloc) {
		out, err = os.OpenFile(fileloc, os.O_RDWR, os.ModePerm)
		if err != nil {
			return 0, err
		}
	} else {
		out, err = os.Create(fileloc)
		if err != nil {
			return 0, err
		}
	}
	defer out.Close()

	n, err := io.Copy(out, get.Body)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// FileExists returns whether or not the file exists on the disk
func FileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}
