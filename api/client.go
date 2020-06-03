package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/deta/deta-cli/auth"
)

const (
	rootEndpoint = "https://web.deta.sh"
)

// set with Makefile during compilation
var version string

// DetaClient client that talks with the deta api
type DetaClient struct {
	rootEndpoint string
	client       *http.Client
}

// NewDetaClient new client to talk with the deta api
func NewDetaClient() *DetaClient {
	var e string
	if version == "DEV" {
		e = os.Getenv("DEV_ENDPOINT")
	} else {
		e = rootEndpoint
	}
	return &DetaClient{
		rootEndpoint: e,
		client:       &http.Client{},
	}
}

// RequestInput input to Request function
type RequestInput struct {
	Path        string
	Method      string
	Headers     map[string]string
	QueryParams map[string]string
	Body        interface{}
	NeedsAuth   bool
	ContentType string
}

// RequestOutput ouput of Request function
type RequestOutput struct {
	Status int
	Header http.Header
	Body   interface{}
}

// Request send an http request to the deta api
func (d *DetaClient) Request(i *RequestInput) (*RequestOutput, error) {
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
		token, err := authManager.GetAccessToken()
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authoriazation", fmt.Sprintf("Bearer %s", token))
	}

	// headers
	if i.ContentType != "" {
		req.Header.Set("Content-type", i.ContentType)
	} else {
		// default set to application/json
		req.Header.Set("Content-type", "application/json")
	}
	for k, v := range i.Headers {
		req.Header.Set(k, v)
	}

	// query params
	q := req.URL.Query()
	for k, v := range i.QueryParams {
		q.Add(k, v)
	}

	res, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var responseBody interface{}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &responseBody)
	if err != nil {
		return nil, err
	}

	o := &RequestOutput{
		Status: res.StatusCode,
		Header: res.Header,
		Body:   responseBody,
	}
	return o, nil
}
