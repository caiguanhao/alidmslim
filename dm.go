package alidmslim

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	accountName     string
	accessKeyId     string
	accessKeySecret string
	debug           bool
}

// NewClient creates a client given access key ID and secret.
func NewClient(accountName, id, secret string) *Client {
	return &Client{
		accountName:     accountName,
		accessKeyId:     id,
		accessKeySecret: secret,
		debug:           false,
	}
}

// Creates a new client with debug (more information printed to stderr).
func (client *Client) Debug(debug bool) *Client {
	return &Client{
		accountName:     client.accountName,
		accessKeyId:     client.accessKeyId,
		accessKeySecret: client.accessKeySecret,
		debug:           debug,
	}
}

type Mail struct {
	subject string
	content string
	isHtml  bool
	client  *Client
}

// NewMail creates a new mail given subject and text content.
func (client *Client) NewMail(subject, content string) *Mail {
	return &Mail{
		client:  client,
		subject: subject,
		content: content,
		isHtml:  false,
	}
}

// NewHTMLMail creates a new mail given subject and html content.
func (client *Client) NewHTMLMail(subject, content string) *Mail {
	return &Mail{
		client:  client,
		subject: subject,
		content: content,
		isHtml:  true,
	}
}

// MustSend sends a mail to target addresses, panics if send operation fails.
func (mail Mail) MustSend(ctx context.Context, addresses ...string) {
	if err := mail.Send(ctx, addresses...); err != nil {
		panic(err)
	}
}

// Send sends a mail to target addresses.
func (mail Mail) Send(ctx context.Context, addresses ...string) error {
	if len(addresses) == 0 {
		return errors.New("need at least one email address")
	}
	v := url.Values{}
	v.Set("Format", "json")
	v.Set("Version", "2015-11-23")
	v.Set("AccessKeyId", mail.client.accessKeyId)
	v.Set("SignatureMethod", "HMAC-SHA1")
	v.Set("Timestamp", time.Now().UTC().Format(time.RFC3339))
	v.Set("SignatureVersion", "1.0")
	v.Set("SignatureNonce", randomString(64))
	v.Set("Action", "SingleSendMail")
	v.Set("AccountName", mail.client.accountName)
	v.Set("ReplyToAddress", "false")
	v.Set("AddressType", "0")
	v.Set("Subject", mail.subject)
	if mail.isHtml {
		v.Set("HtmlBody", mail.content)
	} else {
		v.Set("TextBody", mail.content)
	}
	v.Set("ToAddress", strings.Join(addresses, ","))
	h := hmac.New(sha1.New, []byte(mail.client.accessKeySecret+"&"))
	h.Write([]byte("POST&%2F&" + urlEncode(v.Encode())))
	v.Set("Signature", base64.StdEncoding.EncodeToString(h.Sum(nil)))
	req, err := http.NewRequestWithContext(ctx, "POST", "https://dm.aliyuncs.com/", strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	if mail.client.debug {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return err
		}
		log.Println(string(dump))
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if mail.client.debug {
		dumpBody := strings.Contains(resp.Header.Get("Content-Type"), "json")
		dump, err := httputil.DumpResponse(resp, dumpBody)
		if err != nil {
			return err
		}
		log.Println(string(dump))
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var rerr ResponseError
	json.Unmarshal(b, &rerr)
	if rerr.Code != "" {
		return rerr
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("returned status %d instead of 200", resp.StatusCode)
	}
	return nil
}

type ResponseError struct {
	Message string
	Code    string
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("%s Error: %s", e.Code, e.Message)
}

func urlEncode(input string) string {
	return url.QueryEscape(strings.Replace(strings.Replace(strings.Replace(input, "+", "%20", -1), "*", "%2A", -1), "%7E", "~", -1))
}

func randomString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
