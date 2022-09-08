package browse

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func List(client *http.Client, year, month int) ([]byte, error) {
	u, _ := url.Parse("https://moneyforward.com")
	c := http.Cookie{
		Name:   "cf_last_fetch_from_date",
		Value:  fmt.Sprintf("%04d/%02d/%02d", year, month, 1),
		Path:   "/",
		Domain: "moneyforward.com",
	}
	client.Jar.SetCookies(u, []*http.Cookie{&c})

	u, _ = url.Parse("https://moneyforward.com/cf")
	req, err := newGetRequest(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to build request to /cf: %w", err)
	}

	// dump, err := httputil.DumpRequestOut(req, true)
	// log.Print(string(dump))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to GET /cf: %w", err)
	}
	defer resp.Body.Close()
	time.Sleep(time.Second)

	html, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to build goquery Document: %w", err)
	}

	date, err := doc.Find(".date_range h2").Html()
	if err != nil {
		return nil, fmt.Errorf("failed to check date")
	}

	if !strings.Contains(date, fmt.Sprintf("%04d/%02d", year, month)) {
		return nil, nil
	}

	return html, nil
}
