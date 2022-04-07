package wrikedaemon

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const wrikeLoginURL = "https://login.wrike.com/oauth2/token"

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Host         string `json:"host"`
	validBefore  time.Time
}

type Daemon struct {
	clientID     string
	clientSecret string
	authCode     string
	authCodeChan chan string

	client       *http.Client
	mut          sync.RWMutex
	currentToken *Token
}

func NewDaemon(clientID, clientSecret string) *Daemon {
	return &Daemon{
		clientID:     clientID,
		clientSecret: clientSecret,
		authCodeChan: make(chan string, 1),
		client:       &http.Client{},
	}
}

func (d *Daemon) Get() (*Token, error) {

	d.mut.RLock()
	defer d.mut.RUnlock()

	if d.currentToken == nil {
		return nil, errors.New("token not ready yet")
	}

	if time.Now().After(d.currentToken.validBefore) {
		return nil, errors.New("token expired")
	}

	return d.currentToken, nil
}

func (d *Daemon) SetAuthCode(authCode string) {
	d.authCodeChan <- authCode
}

func (d *Daemon) StartPolling() {

polling:
	log.Printf("waiting for auth code...")
	code := <-d.authCodeChan
	d.authCode = code

	log.Printf("getting access token...")
	token, err := d.resolveToken()
	if err != nil {
		log.Printf("cannot resolve token from wrike authentication API: %v", err)
		goto polling
	}

	d.mut.Lock()
	d.currentToken = token
	d.mut.Unlock()

	timer := time.NewTimer(time.Duration(token.ExpiresIn) * time.Second)
	for {
		select {
		case <-timer.C:
			token, err = d.refreshToken()
			if err != nil {
				log.Printf("cannot resolve refresh token: %v", err)
				continue
			}

			d.mut.Lock()
			d.currentToken = token
			d.mut.Unlock()
		}
	}
}

func (d *Daemon) resolveToken() (*Token, error) {

	requestValues := make(url.Values)
	requestValues.Set("client_id", d.clientID)
	requestValues.Set("client_secret", d.clientSecret)
	requestValues.Set("grant_type", "authorization_code")
	requestValues.Set("code", d.authCode)

	res, err := d.client.PostForm(wrikeLoginURL, requestValues)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("cannot close body of access token response: %v", err)
		}
	}()

	log.Printf("access response status: %s", res.Status)

	if res.StatusCode != http.StatusOK {
		if b, err := ioutil.ReadAll(res.Body); err != nil {
			log.Printf("cannot read response body: %v", err)
		} else {
			log.Printf("response body:\n%s", string(b))
		}

		return nil, errors.New("wrike returned a non-200 response")
	}

	tok := new(Token)
	if err := json.NewDecoder(res.Body).Decode(tok); err != nil {
		return nil, err
	}

	tok.validBefore = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
	return tok, nil
}

func (d *Daemon) refreshToken() (*Token, error) {

	tok, err := d.Get()
	if err != nil {
		return nil, err
	}

	requestValues := make(url.Values)
	requestValues.Set("client_id", d.clientID)
	requestValues.Set("client_secret", d.clientSecret)
	requestValues.Set("grant_type", "refresh_token")
	requestValues.Set("refresh_token", tok.RefreshToken)

	res, err := d.client.PostForm(wrikeLoginURL, requestValues)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("cannot close body of access token response: %v", err)
		}
	}()

	log.Printf("refresh response status: %s", res.Status)

	if res.StatusCode != http.StatusOK {
		if b, err := ioutil.ReadAll(res.Body); err != nil {
			log.Printf("cannot read response body: %v", err)
		} else {
			log.Printf("response body:\n%s", string(b))
		}

		return nil, errors.New("wrike returned a non-200 response")
	}

	tok = new(Token)
	if err := json.NewDecoder(res.Body).Decode(tok); err != nil {
		return nil, err
	}

	tok.validBefore = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
	return tok, nil
}
