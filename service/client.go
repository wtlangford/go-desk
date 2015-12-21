package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nhjk/oauth"
	desk "github.com/wtlangford/go-desk"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	client       *http.Client
	BaseURL      *url.URL
	userEmail    string
	userPassword string
	consumer     oauth.Consumer
	token        oauth.Token
	useOAuth     bool
	Case         *CaseService
	Customer     *CustomerService
	Company      *CompanyService
	User         *UserService
	Group        *GroupService
	Job          *JobService
	MaxRetries   int
}

func NewClient(httpClient *http.Client, endpointURL string, userEmail string, userPassword string) *Client {
	cli := newClient(httpClient, endpointURL)
	cli.UseBasicAuth(userEmail, userPassword)
	return cli
}

func NewClientWithOAuth(httpClient *http.Client, endpointURL, consumerKey, consumerSecret, tokenKey, tokenSecret string) *Client {
	cli := newClient(httpClient, endpointURL)
	cli.UseOAuth(consumerKey, consumerSecret, tokenKey, tokenSecret)
	return cli
}

func newClient(httpClient *http.Client, endpointURL string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(fmt.Sprintf("%s/api/%s/", endpointURL, desk.DeskApiVersion))
	c := &Client{client: httpClient, BaseURL: baseURL}
	c.Case = NewCaseService(c)
	c.Customer = &CustomerService{client: c}
	c.Company = &CompanyService{client: c}
	c.User = &UserService{client: c}
	c.Group = &GroupService{client: c}
	c.Job = &JobService{client: c}
	c.MaxRetries = -1
	return c
}

func (c *Client) UseOAuth(consumerKey, consumerSecret, tokenKey, tokenSecret string) {
	c.consumer = oauth.Consumer{Key: consumerKey, Secret: consumerSecret}
	c.token = oauth.Token{Key: tokenKey, Secret: tokenSecret}
	c.useOAuth = true
}

func (c *Client) UseBasicAuth(userEmail, userPassword string) {
	c.userEmail = userEmail
	c.userPassword = userPassword
	c.useOAuth = false
}

func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)

	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	buf := new(bytes.Buffer)
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
		b, err := json.MarshalIndent(body, "", "  ")
		if err == nil {
			log.Printf("%s %s [request]\n%s", method, u.String(), b)
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if c.useOAuth {
		c.consumer.Authorize(req, &c.token)
	} else {
		req.SetBasicAuth(c.userEmail, c.userPassword)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", desk.DeskUserAgent)
	return req, nil
}

// Do sends an API request and returns the API response.  The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred.  If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	log.Printf("Do %v", req)

	var resp *http.Response
	var err error

	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		return nil, err
	}

	var retries = c.MaxRetries
	limitRetries := retries >= 0
	for {
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		resp, err = c.client.Do(req)

		if err != nil {
			return nil, err
		}

		// if we get a 429 response code we should try the request again
		// otherwise we simply break out of the loop and continue
		if resp.StatusCode != 429 {
			break
		}

		retries -= 1
		if limitRetries && retries < 0 {
			break
		}

		// resp will be overwritten so close the body before that happens
		resp.Body.Close()

		// get the amount of time till the rate limit reset will occur
		nextWindow, _ := strconv.Atoi(resp.Header.Get("X-Rate-Limit-Reset"))

		// sleep till the rate limit has been reset then continue in the loop
		// so the request gets retried
		time.Sleep(time.Second * time.Duration(nextWindow))
	}

	defer resp.Body.Close()

	err = CheckResponse(resp)

	if err != nil {
		return resp, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err == nil {
				b, indentErr := json.MarshalIndent(v, "", "  ")
				if indentErr == nil {
					log.Printf("%s %v [response]\n%s", req.Method, req.URL, b)
				}
			}
		}
	}
	return resp, err
}

type ErrorResponse struct {
	Response *http.Response
	Errors   map[string]interface{} `json:"errors"`
	Message  string                 `json:"message"`
}

func (r *ErrorResponse) Error() string {
	errstr, _ := json.Marshal(r.Errors)
	return fmt.Sprintf("%v %v: %d %v: %v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Message, string(errstr))
}

func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}
	return errorResponse
}
