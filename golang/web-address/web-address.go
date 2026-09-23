// Package webaddress implements the url builder
package webaddress

import (
	"net/url"
)

// WebAddress implements the builder for the web-address creation
// valid:- https:://https://graph.facebook.com/oauth/access_token
// invalid:- https://graph.facebook.com/oauth/access_token/
type WebAddress struct {
	base   *url.URL
	client *client
}

func New(baseURL string) *WebAddress {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}
	return &WebAddress{
		base:   u,
		client: newClient(baseURL),
	}
}

func (w *WebAddress) Field(key, value string) *WebAddress {
	q := w.base.Query()
	q.Add(key, value)
	w.base.RawQuery = q.Encode()
	return w
}

// Path appends a segment to the URL path
// e.g., calling .Path("me") on "https://graph.facebook.com/v25.0"
// results in "https://graph.facebook.com/v25.0/me"
func (w *WebAddress) Path(segment string) *WebAddress {
	// Ensure there is a slash between segments
	if w.base.Path == "" || w.base.Path[len(w.base.Path)-1] != '/' {
		w.base.Path += "/"
	}
	w.base.Path += segment
	return w
}
func (w *WebAddress) Delete(key string) *WebAddress {
	q := w.base.Query()
	q.Del(key)
	w.base.RawQuery = q.Encode()
	return w
}

func (w *WebAddress) Generate() string {
	return w.base.String()
}

func (w *WebAddress) Request() *client {
	return w.client
}
