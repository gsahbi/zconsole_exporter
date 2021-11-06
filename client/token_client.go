package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"zconsole_exporter/util/config"

	"github.com/google/go-querystring/query"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type authToken struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type tokenClient struct {
	tgt  url.URL
	hc   HTTPClient
	ctx  context.Context
	tok  config.Token
	atok authToken
}

func (c *tokenClient) authenticate() error {
	u := c.tgt
	u.Path = "api/auth/v1/api_keys/login"

	values := map[string]string{"clientId": c.tok.ClientID, "secret": c.tok.AuthToken}
	jsonValue, _ := json.Marshal(values)

	r, _ := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonValue))
	r.Header.Add("Content-Type", "application/json")

	resp, err := c.hc.Do(r)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("response code was %d, expected 200 (path: %q)", resp.StatusCode, u.Path)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &c.atok)

}

func (c *tokenClient) newGetRequest(url string) (*http.Request, error) {
	r, err := http.NewRequestWithContext(c.ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.atok.AccessToken == "" {
		err = c.authenticate()
		if err != nil {
			return nil, err
		}
	}

	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.atok.AccessToken))
	return r, nil
}

func (c *tokenClient) Get(path string, q interface{}, obj interface{}, raw *string) error {
	u := c.tgt
	u.Path = path

	v, _ := query.Values(q)
	u.RawQuery = v.Encode()

	req, err := c.newGetRequest(u.String())
	if err != nil {
		return err
	}

	req = req.WithContext(c.ctx)
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("response code was %d, expected 200 (path: %q)", resp.StatusCode, path)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if raw != nil {
		*raw = string(b)
		return nil
	}
	return json.Unmarshal(b, obj)
	
}

func (c *tokenClient) String() string {
	return c.tgt.String()
}

func newTokenClient(ctx context.Context, tgt url.URL, hc HTTPClient, token config.Token) (*tokenClient, error) {
	return &tokenClient{tgt, hc, ctx, token, authToken{"", ""}}, nil
}
