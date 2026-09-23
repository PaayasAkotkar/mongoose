// Package webaddress implements the url builder and a small grouped HTTP client
package webaddress

import (
	cheetah "app/mongoose/cheetha"
	"app/mongoose/stack"
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

// reqEntry pairs a request with its eventual result
type reqEntry struct {
	req *http.Request
	ret *Result
}

// client groups requests under a string key so related requests
type client struct {
	base string

	reqs map[string]map[string]*reqEntry

	keys *stack.Stack[string]

	ids map[string]*stack.Stack[string]

	cheetah *cheetah.Cheetah[string, Result]
	mu      sync.Mutex
}

type Result struct {
	Error   error
	IsReady bool
	Result  []byte
}

func genID() string {
	return uuid.New().String()
}

func newClient(url string) *client {
	return &client{
		base:    url,
		reqs:    make(map[string]map[string]*reqEntry),
		keys:    stack.NewStack[string](),
		ids:     make(map[string]*stack.Stack[string]),
		cheetah: cheetah.New[string, Result](100),
	}
}

func (c *client) SetBase(url string) *client {
	c.base = url
	return c
}

// Add registers a new request under the given key (string).
// Multiple calls with the same key group requests together.
func (c *client) Add(key string, method string, load []byte) *client {
	c.mu.Lock()
	defer c.mu.Unlock()
	req, err := http.NewRequest(method, c.base, bytes.NewBuffer(load))
	if err != nil {
		log.Println(err)
		return c
	}

	id := genID()

	if _, ok := c.reqs[key]; !ok {
		c.reqs[key] = make(map[string]*reqEntry)
		c.ids[key] = stack.NewStack[string]()
		c.keys.Push(key)
	}

	c.reqs[key][id] = &reqEntry{req: req, ret: &Result{}}
	c.ids[key].Push(id)
	return c
}

// currentKey returns the most recently added type-key, or false if none exist.
func (c *client) currentKey() (string, bool) {
	k := c.keys.Latest()
	if k == nil {
		return "", false
	}
	return *k, true
}

// SetHeader applies a header to every request in the most recently added group.
func (c *client) SetHeader(key, value string) *client {
	tk, ok := c.currentKey()
	if !ok {
		return c
	}
	for _, entry := range c.reqs[tk] {
		entry.req.Header.Add(key, value)
	}
	return c
}

// Release empties the fields
func (c *client) Release() *client {
	c.mu.Lock()
	defer c.mu.Unlock()
	for id := range c.reqs {
		delete(c.reqs, id)
		delete(c.ids, id)
		c.keys.Erase(id)
	}
	return c
}

func (c *client) DelHeader(key string) *client {
	tk, ok := c.currentKey()
	if !ok {
		return c
	}
	for _, entry := range c.reqs[tk] {
		entry.req.Header.Del(key)
	}
	return c
}

// Remove deletes an entire group of requests by its string key.
func (c *client) Remove(key string) *client {
	delete(c.reqs, key)
	delete(c.ids, key)
	c.keys.Erase(key)
	return c
}

// Delete removes a single request (by id) from the group identified by key.
func (c *client) Delete(key string, id string) *client {
	if group, ok := c.reqs[key]; ok {
		delete(group, id)
	}
	if idStack, ok := c.ids[key]; ok {
		idStack.Erase(id)
	}
	return c
}

// doRequest peforms io read and write
func (c *client) doRequest(id string, entry *reqEntry) {
	httpClient := &http.Client{}
	resp, err := httpClient.Do(entry.req)
	if err != nil {
		entry.ret.Error = err
		c.cheetah.Publish(id, entry.ret)
		return
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	entry.ret.Result = b
	entry.ret.IsReady = true
	c.cheetah.Publish(id, entry.ret)
}

// Go manually firing every pending request synchronously across all groups.
func (c *client) Go() {
	for _, group := range c.reqs {
		for id, entry := range group {
			c.doRequest(id, entry)
		}
	}
}

type Handler func(result *Result)

// GoMonitor fires every request concurrently and streams results to handler.
// no need to use Go method if used this
func (c *client) GoMonitor(ctx context.Context, handler Handler, fns ...func()) {
	out := make(chan *Result, 100)

	for _, fn := range fns {
		fn()
	}

	go func() {
		for {
			select {
			case result, ok := <-out:
				if !ok {
					return
				}
				if result == nil {
					continue
				}
				handler(result)
			case <-ctx.Done():
				return
			}
		}
	}()

	for key, group := range c.reqs {
		for id, entry := range group {
			id, entry := id, entry
			go c.doRequest(id, entry)
			go func(id string) {
				ch := c.cheetah.Subscribe(id)
				defer c.cheetah.Unsubscribe(id, ch)

				select {
				case <-ctx.Done():
					return
				case result, ok := <-ch:
					if !ok || result == nil {
						return
					}
					select {
					case out <- result:
					case <-ctx.Done():
					}
				}
			}(id)
		}
		_ = key
	}
}
