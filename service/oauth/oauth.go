// RFC 5849 Oauth 1 Authorization implementation

// This file originated from github.com/nhjk/oauth, and has been copied in
// after the repository went offline

/*

The MIT License (MIT)

Copyright (c) 2014 Nikias Kalpaxis

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.

*/

package oauth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Consumer struct {
	Key, Secret string
}

type Token struct {
	Key, Secret string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Authorize turns a request into valid oauth request by adding the
// "Authorization" header
func (c *Consumer) Authorize(req *http.Request, t *Token) {
	authorization := fmt.Sprintf("OAuth "+
		`oauth_consumer_key="%s",`+
		`oauth_token="%s",`+
		`oauth_signature_method="%s",`+
		`oauth_timestamp="%s",`+
		`oauth_nonce="%s",`+
		`oauth_version="%s"`,
		c.Key, t.Key, signatureMethod, timestamp(), nonce(), "1.0")

	req.Header.Set("Authorization", authorization)
	authorization += `,oauth_signature="` + c.Signature(req, t) + `"`

	req.Header.Set("Authorization", authorization)
}

const signatureMethod = "HMAC-SHA1"

func timestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

const nonceLength = 20

var alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func nonce() string {
	nonce := make([]byte, nonceLength)
	for i := 0; i < nonceLength; i++ {
		nonce[i] = alphanumeric[rand.Intn(len(alphanumeric))]
	}
	return string(nonce)
}

// Signature returns the oauth signature of an http.Request that has valid
// oauth header parameters
func (c *Consumer) Signature(req *http.Request, t *Token) string {
	message := req.Method + "&" + encode(baseUri(req.URL)) + "&" + encode(requestParameters(req))
	key := encode(c.Secret) + "&" + encode(t.Secret)

	// encrypt with HMAC-SHA1
	hash := hmac.New(sha1.New, []byte(key))
	hash.Write([]byte(message))
	hmacSig := hash.Sum(nil)

	// base64 encode
	base64Sig := make([]byte, base64.StdEncoding.EncodedLen(len(hmacSig)))
	base64.StdEncoding.Encode(base64Sig, hmacSig)

	// percent encode and return
	return encode(string(base64Sig))
}

func baseUri(uri *url.URL) string {
	return strings.Split(uri.String(), "?")[0]
}

// Implements percent encoding. The go std library implementation of
// url.QueryEscape is not valid for the oauth spec. Mainly spaces gettting
// encoded as "+" instead of "%20"
func encode(s string) string {
	e := []byte(nil)
	for i := 0; i < len(s); i++ {
		b := s[i]
		if encodable(b) {
			e = append(e, '%')
			e = append(e, "0123456789ABCDEF"[b>>4])
			e = append(e, "0123456789ABCDEF"[b&15])
		} else {
			e = append(e, b)
		}
	}
	return string(e)
}

func encodable(b byte) bool {
	return !('A' <= b && b <= 'Z' || 'a' <= b && b <= 'z' ||
		'0' <= b && b <= '9' || b == '-' || b == '.' || b == '_' || b == '~')
}

func requestParameters(req *http.Request) string {
	var parameters []*parameter

	// Add query string parameters to parameters
	query := map[string][]string(req.URL.Query())
	for name, values := range query {
		for _, value := range values {
			p := &parameter{name, value}
			parameters = append(parameters, p)
		}
	}

	// Add form encoded body values to parameters
	if req.ContentLength > 0 && req.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		// read body
		body, _ := ioutil.ReadAll(req.Body)

		// parse body and add params to parameters
		values, _ := url.ParseQuery(string(body))
		values = map[string][]string(values)
		for name, vs := range values {
			for _, value := range vs {
				p := &parameter{name, value}
				parameters = append(parameters, p)
			}
		}

		// reset body
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	// Add the oauth_* fields and their values to parameters
	header := strings.Join(req.Header["Authorization"], ",")
	submatches := oauthFieldRegexp.FindAllStringSubmatch(header, -1)
	for i := 0; i < len(submatches); i++ {
		// error if len(sumatches[i]) != 3
		if submatches[i][1] != "oauth_signature" {
			p := &parameter{submatches[i][1], submatches[i][2]}
			parameters = append(parameters, p)
		}
	}

	// RFC 5849 3.4.1.3.2. Parameters Normalization
	// 1. encode
	for _, parameter := range parameters {
		parameter.encode()
	}

	// 2. sort
	sort.Sort(byNameThenValue(parameters))

	// 3. concatenate name and value
	concatenatedPairs := make([]string, len(parameters))
	for i, parameter := range parameters {
		concatenatedPairs[i] = parameter.name + "=" + parameter.value
	}

	// 4. concatenate parameters
	var concatenated string
	for _, pair := range concatenatedPairs {
		concatenated += "&" + pair
	}

	// take off first "&"
	return concatenated[1:]
}

// Gets the names and values of the oauth_* parameters in the "Authorization"
// header
var oauthFieldRegexp = regexp.MustCompile(`(oauth_\w+)="(.+?)"`)

// For storing the parameters needed to create an oauth signature
type parameter struct {
	name  string
	value string
}

func (p *parameter) encode() {
	p.name = encode(p.name)
	p.value = encode(p.value)
}

// Implements sorting of parameters by name, then value
type byNameThenValue []*parameter

func (a byNameThenValue) Len() int      { return len(a) }
func (a byNameThenValue) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byNameThenValue) Less(i, j int) bool {
	if a[i].name == a[j].name {
		return a[i].value < a[j].value
	}
	return a[i].name < a[j].name
}
