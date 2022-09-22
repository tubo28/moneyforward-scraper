package browse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	embedJSONRegexp = regexp.MustCompile(`gon\.authorizationParams=({.*?})`)
)

type embedJSON struct {
	ClientID     string `json:"clientId"`
	RedirectURI  string `json:"redirectUri"`
	ResponseType string `json:"responseType"`
	Scope        string `json:"scope"`
	State        string `json:"state"`
	Nonce        string `json:"nonce"`
}

func CheckLogin(client *http.Client) (bool, error) {
	req, _ := newGetRequest("https://moneyforward.com/profile")
	resp, err := client.Do(req)
	time.Sleep(time.Second)
	if err != nil {
		return false, fmt.Errorf("failed to load /profile: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read body of /profile: %w", err)
	}

	return bytes.Contains(body, []byte("アカウント設定")), nil
}

func Login(client *http.Client, mail, password string) error {
	var content url.Values

	// 履歴一覧画面
	{
		req, _ := newGetRequest("https://moneyforward.com/cf")
		resp, err := client.Do(req)
		time.Sleep(time.Second)
		if err != nil {
			return fmt.Errorf("failed to load /cf: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read body of /cf: %w", err)
		}

		jsMatched := embedJSONRegexp.FindSubmatch(body)
		if jsMatched == nil {
			// ログイン済み
			log.Print("Login called, but it seems already logged in for ", mail)
			return nil
		}
		var js embedJSON
		if err := json.Unmarshal(jsMatched[1], &js); err != nil {
			return fmt.Errorf("failed to read body of /cf: %w", err)
		}

		content = url.Values{}
		content.Add("client_id", js.ClientID)
		content.Add("nonce", js.Nonce)
		content.Add("redirect_uri", js.RedirectURI)
		content.Add("response_type", js.ResponseType)
		content.Add("scope", js.Scope)
		content.Add("state", js.State)

	}

	// メールアドレス入力画面
	{
		u, err := url.Parse("https://id.moneyforward.com/sign_in/email")
		if err != nil {
			return fmt.Errorf("failed to parse url to /sign_in/email: %w", err)
		}
		u.RawQuery = content.Encode()

		req, _ := newGetRequest(u.String())
		resp, err := client.Do(req)
		time.Sleep(time.Second)
		if err != nil {
			return fmt.Errorf("failed to GET /sign_in/email: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to load body of /sign_in/email: %w", err)
		}

		jsMatched := embedJSONRegexp.FindSubmatch(body)
		if jsMatched == nil {
			return fmt.Errorf("failed to parse params in login page: %w", err)
		}
		var js embedJSON
		if json.Unmarshal(jsMatched[1], &js); err != nil {
			return fmt.Errorf("failed to parse json in login page: %w", err)
		}

		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("cannot build goquery Document: %w", err)
		}

		csrfToken, exists := doc.Find("meta[name='csrf-token']").Attr("content")
		if !exists {
			return fmt.Errorf("csrf-token is not set: %w", err)
		}

		content = url.Values{}
		content.Add("authenticity_token", csrfToken)
		content.Add("_method", "post")
		content.Add("client_id", js.ClientID)
		content.Add("redirect_uri", js.RedirectURI)
		content.Add("response_type", js.ResponseType)
		content.Add("scope", js.Scope)
		content.Add("state", js.State)
		content.Add("nonce", js.Nonce)
		content.Add("mfid_user[email]", mail)
		content.Add("hiddenPassword", "")
	}

	// パスワード入力画面
	{
		req, _ := newPostFormRequest("https://id.moneyforward.com/sign_in/email", content)
		resp, err := client.Do(req)
		time.Sleep(time.Second)
		if err != nil {
			return fmt.Errorf("failed to parse url to /sign_in/email: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to load body of /sign_in/email: %w", err)
		}

		jsMatched := embedJSONRegexp.FindSubmatch(body)
		if jsMatched == nil {
			return fmt.Errorf("failed to parse params in login page: %w", err)
		}
		var js embedJSON
		if json.Unmarshal(jsMatched[1], &js); err != nil {
			return fmt.Errorf("failed to parse json in login page: %w", err)
		}

		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("cannot build goquery Document: %w", err)
		}

		csrfToken, exists := doc.Find("meta[name='csrf-token']").Attr("content")
		if !exists {
			return fmt.Errorf("csrf-token is not set: %w", err)
		}

		content = url.Values{}
		content.Add("authenticity_token", csrfToken)
		content.Add("_method", "post")
		content.Add("client_id", js.ClientID)
		content.Add("redirect_uri", js.RedirectURI)
		content.Add("response_type", js.ResponseType)
		content.Add("scope", js.Scope)
		content.Add("state", js.State)
		content.Add("nonce", js.Nonce)
		content.Add("mfid_user[email]", mail)
		content.Add("mfid_user[password]", password)
	}

	{

		req, _ := newPostFormRequest("https://id.moneyforward.com/sign_in", content)
		resp, err := client.Do(req)
		time.Sleep(time.Second)
		if err != nil {
			return fmt.Errorf("failed to POST /sign_in: %w", err)
		}
		defer resp.Body.Close()
		io.ReadAll(resp.Body)
	}

	return nil
}
