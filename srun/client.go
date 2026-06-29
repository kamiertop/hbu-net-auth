// Codex+GPT5.5 生成

package srun

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	defaultTimeout = 15 * time.Second
	portalPath     = "/cgi-bin/srun_portal"
	challengePath  = "/cgi-bin/get_challenge"
)

type Options struct {
	PortalURL string
	APIBase   string
	CheckURL  string
}

type Client struct {
	portalURL string
	apiBase   string
	checkURL  string
	http      *http.Client
}

type portalVars struct {
	acid string
	ip   string
}

type Response struct {
	raw string
}

func NewClient(options Options) *Client {
	return &Client{
		portalURL: options.PortalURL,
		apiBase:   strings.TrimRight(options.APIBase, "/"),
		checkURL:  options.CheckURL,
		http:      &http.Client{Timeout: defaultTimeout},
	}
}

func (c *Client) Login(username, password string) (Response, error) {
	vars, err := c.fetchPortalVars()
	if err != nil {
		return Response{}, err
	}

	token, err := c.getChallenge(username, vars)
	if err != nil {
		return Response{}, err
	}

	hmd5 := hmacMD5(password, token)
	info := encodeUserInfo(username, password, vars.ip, vars.acid, token)
	chkstr := token + username +
		token + hmd5 +
		token + vars.acid +
		token + vars.ip +
		token + srunN +
		token + srunType +
		token + info

	body, err := c.get(c.buildURL(portalPath, map[string]string{
		"callback":     "hbunet",
		"action":       "login",
		"username":     username,
		"password":     "{MD5}" + hmd5,
		"os":           "Linux",
		"name":         "Linux",
		"double_stack": "1",
		"chksum":       sha1Hex(chkstr),
		"info":         info,
		"ac_id":        vars.acid,
		"ip":           vars.ip,
		"n":            srunN,
		"type":         srunType,
		"_":            timestampMS(),
	}))
	if err != nil {
		return Response{}, err
	}

	return Response{raw: jsonpPayload(body)}, nil
}

func (c *Client) Logout(username string) (Response, error) {
	vars, err := c.fetchPortalVars()
	if err != nil {
		return Response{}, err
	}

	body, err := c.get(c.buildURL(portalPath, map[string]string{
		"callback": "hbunet",
		"action":   "logout",
		"username": username,
		"ip":       vars.ip,
		"ac_id":    vars.acid,
		"_":        timestampMS(),
	}))
	if err != nil {
		return Response{}, err
	}

	return Response{raw: jsonpPayload(body)}, nil
}

// CheckNetwork checks the network connectivity by sending HTTP GET requests to the specified check URLs.
func (c *Client) CheckNetwork() error {
	fmt.Println("请求：", c.checkURL)

	resp, err := c.http.Get(c.checkURL)
	if err != nil {
		fmt.Println("网络不可用：", err)
		return err
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}

	fmt.Printf("网络可达，HTTP %d\n", resp.StatusCode)
	return nil
}

func (c *Client) fetchPortalVars() (portalVars, error) {
	body, err := c.get(c.portalURL)
	if err != nil {
		return portalVars{}, err
	}

	acid := extractJSString(body, "acid")
	if acid == "" {
		acid = "1"
	}
	ip := extractJSString(body, "ip")
	if ip == "" {
		return portalVars{}, fmt.Errorf("未能从 portal 页面解析当前 IP")
	}

	return portalVars{acid: acid, ip: ip}, nil
}

func (c *Client) getChallenge(username string, vars portalVars) (string, error) {
	body, err := c.get(c.buildURL(challengePath, map[string]string{
		"callback": "hbunet",
		"username": username,
		"ip":       vars.ip,
		"_":        timestampMS(),
	}))
	if err != nil {
		return "", err
	}

	payload := jsonpPayload(body)
	value := jsonString(payload, "challenge")
	if value == "" {
		return "", fmt.Errorf("get_challenge 响应中没有 challenge: %s", truncate(body, 300))
	}
	return value, nil
}

func (c *Client) get(rawURL string) (string, error) {
	resp, err := c.http.Get(rawURL)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if err != nil {
		return "", err
	}
	if closeErr != nil {
		return "", closeErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), 300))
	}
	return string(body), nil
}

func (c *Client) buildURL(path string, params map[string]string) string {
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	return c.apiBase + path + "?" + values.Encode()
}

func (r Response) Success() bool {
	return jsonString(r.raw, "error") == "ok" ||
		jsonNumber(r.raw, "code") == 0 && strings.Contains(r.raw, `"code"`) ||
		jsonString(r.raw, "suc_msg") != ""
}

func (r Response) Message() string {
	for _, key := range []string{"suc_msg", "error_msg", "error"} {
		if value := jsonString(r.raw, key); value != "" {
			return value
		}
	}
	return truncate(r.raw, 200)
}

func extractJSString(text, key string) string {
	re := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(key) + `\s*:\s*"([^"]*)"`)
	match := re.FindStringSubmatch(text)
	if len(match) == 2 {
		return match[1]
	}
	return ""
}

func jsonString(payload, key string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return ""
	}
	value, _ := data[key].(string)
	return value
}

func jsonNumber(payload, key string) float64 {
	var data map[string]any
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return -1
	}
	value, ok := data[key].(float64)
	if !ok {
		return -1
	}
	return value
}

func jsonpPayload(value string) string {
	trimmed := strings.TrimSpace(value)
	start := strings.IndexByte(trimmed, '(')
	end := strings.LastIndexByte(trimmed, ')')
	if start >= 0 && end > start {
		return trimmed[start+1 : end]
	}
	return trimmed
}

func timestampMS() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}

func truncate(value string, maxRunes int) string {
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[:maxRunes]) + "..."
}
