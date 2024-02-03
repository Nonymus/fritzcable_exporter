package client

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

type Client struct {
	username string
	password string
	host     string
	sid      string
	refresh  time.Timer
	client   *http.Client
}

type SessionInfo struct {
	SID       string
	Challenge string
	BlockTime int
	Users     []string `xml:"Users>User"`
}

// NewClient creates a new client. To login without a username, supply an empty string
func NewClient(host, username, password string) *Client {
	client := http.Client{
		Timeout: time.Second * 20,
	}
	return &Client{
		password: password,
		username: username,
		host:     host,
		client:   &client,
	}

}

// Login fetches a Session ID for further client interactions
func (c *Client) Login() error {
	loginUrl := fmt.Sprintf("http://%s/login_sid.lua?version=2", c.host)
	res, err := c.client.Get(loginUrl)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var data SessionInfo
	dec := xml.NewDecoder(res.Body)
	err = dec.Decode(&data)
	if err != nil {
		return err
	}
	if data.BlockTime > 0 {
		time.Sleep(time.Duration(data.BlockTime) * time.Second)
	}

	response, err := createResponse(data.Challenge, c.password)
	if err != nil {
		return err
	}
	username := c.username
	if username == "" {
		username = data.Users[0]
	}

	res2, err := c.client.PostForm(loginUrl, url.Values{"username": {username}, "response": {response}})
	if err != nil {
		return err
	}

	var loginData SessionInfo
	defer res2.Body.Close()
	dec = xml.NewDecoder(res2.Body)
	err = dec.Decode(&loginData)
	if err != nil {
		return err
	}
	if loginData.SID == "0000000000000000" {
		return errors.New("did not get a valid session ID")
	}

	c.sid = loginData.SID
	return nil
}

func (c *Client) DocsisStats() (*Data, error) {
	dataUrl := fmt.Sprintf("http://%s/data.lua", c.host)
	values := url.Values{"xhr": {"1"}, "sid": {c.sid}, "lang": {"de"}, "page": {"docInfo"}, "xhrId": {"all"}}
	res, err := c.client.PostForm(dataUrl, values)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	var all Container
	if err := json.Unmarshal(body, &all); err != nil {
		return nil, err
	}

	return &all.Data, nil
}

// CheckLogin validates checks if the currently stored Session ID is still valid
func (c *Client) CheckLogin() (bool, error) {
	loginUrl := fmt.Sprintf("http://%s/login_sid.lua?version=2", c.host)
	values := url.Values{"sid": {c.sid}}
	res, err := c.client.PostForm(loginUrl, values)
	if err != nil {
		return false, err
	}
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	var data SessionInfo
	if err := xml.Unmarshal(body, &data); err != nil {
		return false, err
	}

	return data.SID == c.sid, nil
}

// createResponse computes the PBKDF2-based response to the login challenge by the router
func createResponse(challenge, password string) (string, error) {
	if challenge[0:2] != "2$" {
		return "", errors.New("challenge unsupported")
	}

	parts := strings.Split(challenge, "$")
	if len(parts) != 5 {
		return "", errors.New("challenge missing parameters")
	}

	iter1, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", err
	}

	salt1, err := hex.DecodeString(parts[2])
	if err != nil {
		return "", err
	}

	iter2, err := strconv.Atoi(parts[3])
	if err != nil {
		return "", err
	}

	salt2, err := hex.DecodeString(parts[4])
	if err != nil {
		return "", err
	}

	hash1 := pbkdf2.Key([]byte(password), salt1, iter1, 32, sha256.New)
	hash2 := pbkdf2.Key(hash1, salt2, iter2, 32, sha256.New)

	return fmt.Sprintf("%s$%s", parts[4], hex.EncodeToString(hash2)), nil
}
