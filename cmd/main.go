package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/tubo28/moneyforward-scraper/mf"
	"github.com/tubo28/moneyforward-scraper/mf/browse"
	"github.com/tubo28/moneyforward-scraper/mf/parse"
)

func main() {
	idPassword := os.Getenv("MF_ID_PASSWORD")
	if idPassword == "" {
		panic("no MF_ID_PASSWORD")
	}
	id, password, err := splitIDPassword(idPassword)
	if err != nil {
		panic(err)
	}

	cookie := cookieFileName(id)
	log.Print("cookie file: ", cookie)

	if err := os.MkdirAll(filepath.Dir(cookie), os.ModePerm); err != nil {
		panic(err)
	}
	jar, err := cookiejar.New(&cookiejar.Options{Filename: cookie})
	if err != nil {
		panic(err)
	}

	httpClient := &http.Client{Jar: jar}

	// ログイン確認
	loggedIn, err := browse.CheckLogin(httpClient)
	if err != nil {
		panic(err)
	}

	log.Print("logged in: ", loggedIn)
	if !loggedIn {
		// 明示的にログアウト
		jar.RemoveAll()

		// ログイン
		if err := browse.Login(httpClient, id, password); err != nil {
			panic(err)
		}

		// 再度ログイン確認
		loggedIn, err = browse.CheckLogin(httpClient)
		if err != nil {
			panic(err)
		}
		if !loggedIn {
			log.Fatal("login failed for ", id)
		}
	}

	log.Print("login ok for ", id)
	jar.Save()

	var ret []*mf.MFTransaction

	// 登録までのタイムラグ対策として現在の日付とその14日前の月を取りに行く
	// 今月
	t := time.Now()
	ts, err := fetch(httpClient, id, int(t.Year()), int(t.Month()))
	if err != nil {
		panic(err)
	}
	ret = append(ret, ts...)

	// 先月
	t2 := t.AddDate(0, 0, -14)
	if t.Month() != t2.Month() {
		ts2, err := fetch(httpClient, id, int(t2.Year()), int(t2.Month()))
		if err != nil {
			panic(err)
		}
		ret = append(ret, ts2...)
	}

	sort.Slice(ts, func(i, j int) bool {
		return ret[i].DateRFC3339.Before(ret[j].DateRFC3339) // desc sort
	})
	json.NewEncoder(os.Stdout).Encode(ret)
}

func splitIDPassword(idPassword string) (string, string, error) {
	i := strings.Index(idPassword, "@")
	if i == -1 {
		return "", "", fmt.Errorf("invalid mail:password format")
	}
	j := strings.Index(idPassword[i:], ":")
	if j == -1 {
		return "", "", fmt.Errorf("invalid mail:password format")
	}
	return idPassword[0 : i+j], idPassword[i+j+1:], nil
}

func fetch(client *http.Client, id string, y, m int) ([]*mf.MFTransaction, error) {
	log.Printf("start loading %04d/%02d", y, m)

	html, err := browse.List(client, y, m)
	if err != nil {
		return nil, fmt.Errorf("failed to load list page %04d/%2d: %w", y, m, err)
	}

	ts, err := parse.List(html, id, y, m)
	if err != nil {
		return nil, fmt.Errorf("failed to parse list page %04d/%2d: %w", y, m, err)
	}

	// for _, t := range ts {
	// 	log.Printf("%+v", t)
	// }

	return ts, nil
}

func cookieFileName(id string) string {
	var ret string
	for _, c := range id {
		if !(('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z') || ('0' <= c && c <= '9')) {
			c = '.'
		}
		ret += string(c)
	}
	ret += "_"
	ret += fmt.Sprintf("%x", md5.Sum([]byte(id)))

	home, err := os.UserHomeDir()
	if err != nil {
		// todo
		panic(err)
	}
	ret = filepath.Join(home, "cookie", ret)
	ret, err = filepath.Abs(ret)
	if err != nil {
		panic(err)
	}
	return ret
}
