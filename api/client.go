package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/deta/deta-cli/auth"
)

const (
	rootEndpoint = "https://v1.deta.sh"
	patcherPath  = "patcher"
	viewerPath   = "viewer"
	pigeonPath   = "pigeon"
)

var (
	// set with Makefile during compilation
	version string
)

// DetaClient client that talks with the deta api
type DetaClient struct {
	rootEndpoint string
	client       *http.Client
}

// NewDetaClient new client to talk with the deta api
func NewDetaClient() *DetaClient {
	e := rootEndpoint
	if version == "DEV" {
		fmt.Println("Development mode")
		e = os.Getenv("DEV_ENDPOINT")
		if e == "" {
			os.Stderr.WriteString("Env DEV_ENDPOINT not set\n")
			os.Exit(1)
		}
	}

	return &DetaClient{
		rootEndpoint: e,
		client:       &http.Client{},
	}
}

type errorResp struct {
	Errors  []string `json:"errors,omitempty"`
	Message string   `json:"message,omitempty"`
}

// requestInput input to Request function
type requestInput struct {
	Path        string
	Method      string
	Headers     map[string]string
	QueryParams map[string]string
	Body        interface{}
	NeedsAuth   bool
	ContentType string
}

// requestOutput ouput of Request function
type requestOutput struct {
	Status int
	Body   []byte
	Header http.Header
	Error  *errorResp
}

// Request send an http request to the deta api
func (d *DetaClient) request(i *requestInput) (*requestOutput, error) {
	marshalled, err := json.Marshal(&i.Body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(i.Method, fmt.Sprintf("%s%s", d.rootEndpoint, i.Path), bytes.NewBuffer(marshalled))
	if err != nil {
		return nil, err
	}

	// auth
	if i.NeedsAuth {
		authManager := auth.NewManager()
		tokens, err := authManager.GetTokens()
		if err != nil {
			if os.IsNotExist(err) || errors.Is(err, auth.ErrRefreshTokenInvalid) {
				return nil, fmt.Errorf("login required")
			}
			return nil, fmt.Errorf("failed to authorize")
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokens.AccessToken))
	}

	if i.Body != nil {
		// default set to application/json
		req.Header.Set("Content-type", "application/json")
		if i.ContentType != "" {
			req.Header.Set("Content-type", i.ContentType)
		}
	}

	// headers
	for k, v := range i.Headers {
		req.Header.Set(k, v)
	}

	// query params
	q := req.URL.Query()
	for k, v := range i.QueryParams {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	res, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	o := &requestOutput{
		Status: res.StatusCode,
		Header: res.Header,
	}

	if res.StatusCode >= 200 && res.StatusCode <= 299 && res.StatusCode != 204 {
		if res.StatusCode != 204 {
			o.Body = b
		}
		return o, nil
	}

	var er errorResp
	err = json.Unmarshal(b, &er)
	if err != nil {
		return nil, err
	}
	o.Error = &er
	return o, nil
}
