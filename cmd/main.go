package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/tubo28/moneyforward-scraper/internal/browse"
	"github.com/tubo28/moneyforward-scraper/internal/mf"
	"github.com/tubo28/moneyforward-scraper/internal/parse"
)

func main() {
	id := os.Getenv("MF_ID")
	if id == "" {
		panic("no MF_ID")
	}
	password := os.Getenv("MF_PASSWORD")
	if password == "" {
		panic("no MF_PASSWORD")
	}

	jar, err := cookiejar.New(&cookiejar.Options{Filename: cookieFileName(id)})
	if err != nil {
		panic(err)
	}
	httpClient := &http.Client{Jar: jar}

	if err := browse.Login(httpClient, id, password); err != nil {
		panic(err)
	}
	jar.Save()

	// 登録までのタイムラグ対策として現在の日付とその14日前の月を取りに行く
	t := time.Now()
	fetch(httpClient, int(t.Year()), int(t.Month()))
	t2 := t.AddDate(0, 0, -14)
	if t.Month() != t2.Month() {
		fetch(httpClient, int(t2.Year()), int(t2.Month()))
	}
}

func fetch(client *http.Client, y, m int) ([]*mf.MFTransaction, error) {
	log.Printf("start loading %04d/%02d", y, m)

	html, err := browse.List(client, y, m)
	if err != nil {
		return nil, fmt.Errorf("failed to load list page %04d/%2d: %w", y, m, err)
	}

	ts, err := parse.List(html, y, m)
	if err != nil {
		return nil, fmt.Errorf("failed to parse list page %04d/%2d: %w", y, m, err)
	}

	for _, t := range ts {
		log.Printf("%+v", t)
	}

	return ts, nil
}

func cookieFileName(id string) string {
	ret := cookiejar.DefaultCookieFile()
	ret += "_"
	for _, c := range id {
		if !(('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z') || ('0' <= c && c <= '9')) {
			c = '.'
		}
		ret += string(c)
	}
	ret += "_"
	ret += fmt.Sprintf("%x", md5.Sum([]byte(id)))
	return ret
}
