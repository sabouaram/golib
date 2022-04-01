/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package request

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	libtls "github.com/nabbar/golib/certificates"

	liberr "github.com/nabbar/golib/errors"
)

type request struct {
	s sync.Mutex

	o *atomic.Value
	f FctHttpClient
	u *url.URL
	h url.Values
	p url.Values
	b io.Reader
	m string
	e *requestError
}

func (r *request) _GetDefaultTLS() libtls.TLSConfig {
	if cfg := r.GetOption(); cfg != nil {
		return cfg._GetDefaultTLS()
	}

	return nil
}

func (r *request) _GetClient() *http.Client {
	var h string

	if r.u != nil {
		h = r.u.Hostname()
	}

	if r.f != nil {
		if c := r.f(r._GetDefaultTLS(), h); c != nil {
			return c
		}
	}

	if cfg := r.GetOption(); cfg != nil {
		return cfg.GetClientHTTP(h)
	}

	return &http.Client{}
}

func (r *request) _IsValidCode(listValid []int, statusCode int) bool {
	if len(listValid) < 1 {
		return true
	}

	for _, c := range listValid {
		if c == statusCode {
			return true
		}
	}

	return false
}

func (r *request) Clone() (Request, error) {
	if n, e := r.New(); e != nil {
		return nil, e
	} else {
		r.s.Lock()
		defer r.s.Unlock()

		n.CleanHeader()
		for k := range r.h {
			n.SetHeader(k, r.h.Get(k))
		}

		n.CleanParams()
		for k := range r.p {
			n.SetParams(k, r.p.Get(k))
		}

		return n, nil
	}
}

func (r *request) New() (Request, error) {
	cfg := r.GetOption()

	r.s.Lock()
	defer r.s.Unlock()

	var n *request

	if cfg == nil {
		if i, e := New(r.f, Options{}); e != nil {
			return nil, e
		} else {
			n = i.(*request)
		}
	}

	if r.u != nil {
		n.u = &url.URL{
			Scheme:      r.u.Scheme,
			Opaque:      r.u.Opaque,
			User:        r.u.User,
			Host:        r.u.Host,
			Path:        r.u.Path,
			RawPath:     r.u.RawPath,
			ForceQuery:  r.u.ForceQuery,
			RawQuery:    r.u.RawQuery,
			Fragment:    r.u.Fragment,
			RawFragment: r.u.RawFragment,
		}
	}

	return n, nil
}

func (r *request) GetOption() *Options {
	r.s.Lock()
	defer r.s.Unlock()

	if r.o == nil {
		return nil
	} else if i := r.o.Load(); i == nil {
		return nil
	} else if o, ok := i.(*Options); !ok {
		return nil
	} else {
		return o
	}
}

func (r *request) SetOption(opt *Options) error {
	if e := r.SetEndpoint(opt.Endpoint); e != nil {
		return e
	}

	if opt.Auth.Basic.Enable {
		r.AuthBasic(opt.Auth.Basic.Username, opt.Auth.Basic.Password)
	} else if opt.Auth.Bearer.Enable {
		r.AuthBearer(opt.Auth.Bearer.Token)
	}

	r.s.Lock()
	defer r.s.Unlock()

	if r.o == nil {
		r.o = new(atomic.Value)
	}

	r.o.Store(opt)
	return nil
}

func (r *request) SetClient(fct FctHttpClient) {
	r.s.Lock()
	defer r.s.Unlock()

	r.f = fct
}

func (r *request) SetEndpoint(u string) error {
	if uri, err := url.Parse(u); err != nil {
		return err
	} else {
		r.s.Lock()
		defer r.s.Unlock()

		r.u = uri
		return nil
	}
}

func (r *request) GetEndpoint() string {
	r.s.Lock()
	defer r.s.Unlock()

	return r.u.String()
}

func (r *request) SetPath(raw bool, path string) {
	r.s.Lock()
	defer r.s.Unlock()

	if raw {
		r.u.RawPath = path
	} else {
		r.u.Path = path
	}
}

func (r *request) AddPath(raw bool, path ...string) {
	r.s.Lock()
	defer r.s.Unlock()

	if r.u == nil {
		return
	}

	for i := range path {
		if strings.HasPrefix(path[i], "/") {
			path[i] = strings.TrimPrefix(path[i], "/")
		}

		if strings.HasSuffix(path[i], "/") {
			path[i] = strings.TrimSuffix(path[i], "/")
		}

		if raw {
			r.u.RawPath = filepath.Join(r.u.RawPath, path[i])
		} else {
			r.u.Path = filepath.Join(r.u.Path, path[i])
		}
	}
}

func (r *request) SetMethod(method string) {
	r.s.Lock()
	defer r.s.Unlock()

	switch strings.ToUpper(method) {
	case http.MethodGet:
		r.m = http.MethodGet
	case http.MethodHead:
		r.m = http.MethodHead
	case http.MethodPost:
		r.m = http.MethodPost
	case http.MethodPut:
		r.m = http.MethodPut
	case http.MethodPatch:
		r.m = http.MethodPatch
	case http.MethodDelete:
		r.m = http.MethodDelete
	case http.MethodConnect:
		r.m = http.MethodConnect
	case http.MethodOptions:
		r.m = http.MethodOptions
	case http.MethodTrace:
		r.m = http.MethodTrace
	default:
		r.m = strings.ToUpper(method)
	}

	if r.m == "" {
		r.m = http.MethodGet
	}
}

func (r *request) GetMethod() string {
	r.s.Lock()
	defer r.s.Unlock()

	return r.m
}

func (r *request) CleanParams() {
	r.s.Lock()
	defer r.s.Unlock()

	r.p = make(url.Values)
}

func (r *request) DelParams(key string) {
	r.s.Lock()
	defer r.s.Unlock()

	r.p.Del(key)
}

func (r *request) SetParams(key, val string) {
	r.s.Lock()
	defer r.s.Unlock()

	if len(r.p) < 1 {
		r.p = make(url.Values)
	}

	r.p.Set(key, val)
}

func (r *request) AddParams(key, val string) {
	r.s.Lock()
	defer r.s.Unlock()

	if len(r.p) < 1 {
		r.p = make(url.Values)
	}

	r.p.Set(key, val)
}

func (r *request) GetFullUrl() *url.URL {
	r.s.Lock()
	defer r.s.Unlock()

	return r.u
}

func (r *request) SetFullUrl(u *url.URL) {
	r.s.Lock()
	defer r.s.Unlock()

	r.u = u
}

func (r *request) AuthBearer(token string) {
	r.SetHeader("Authorization", "Bearer "+token)
}

func (r *request) AuthBasic(user, pass string) {
	r.SetHeader("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(user+":"+pass)))
}

func (r *request) ContentType(content string) {
	r.SetHeader("Content-Type", content)
}

func (r *request) CleanHeader() {
	r.s.Lock()
	defer r.s.Unlock()

	r.h = make(url.Values)
}

func (r *request) DelHeader(key string) {
	r.s.Lock()
	defer r.s.Unlock()

	r.h.Del(key)
}

func (r *request) SetHeader(key, value string) {
	r.s.Lock()
	defer r.s.Unlock()

	if len(r.h) < 1 {
		r.h = make(url.Values)
	}

	r.h.Set(key, value)
}

func (r *request) AddHeader(key, value string) {
	r.s.Lock()
	defer r.s.Unlock()

	if len(r.h) < 1 {
		r.h = make(url.Values)
	}

	r.h.Add(key, value)
}

func (r *request) BodyJson(body interface{}) error {
	if p, e := json.Marshal(body); e != nil {
		return e
	} else {
		r.s.Lock()
		defer r.s.Unlock()

		r.b = bytes.NewBuffer(p)
	}

	r.ContentType("application/json")
	return nil
}

func (r *request) BodyReader(body io.Reader, contentType string) {
	r.s.Lock()
	defer r.s.Unlock()

	r.b = body

	if contentType != "" {
		r.ContentType(contentType)
	}
}

func (r *request) Error() RequestError {
	r.s.Lock()
	defer r.s.Unlock()

	return r.e
}

func (r *request) IsError() bool {
	r.s.Lock()
	defer r.s.Unlock()

	return r.e != nil
}

func (r *request) Do(ctx context.Context) (*http.Response, liberr.Error) {
	r.s.Lock()
	defer r.s.Unlock()

	if r.m == "" || r.u == nil || r.u.String() == "" {
		return nil, ErrorParamsInvalid.Error(nil)
	}

	var (
		e   error
		req *http.Request
		rsp *http.Response
		err liberr.Error
	)

	r.e = nil

	req, err = r._MakeRequest(ctx)
	if err != nil {
		return nil, err
	}

	rsp, e = r._GetClient().Do(req)

	if e != nil {
		r.e = &requestError{
			c: 0,
			s: "",
			b: nil,
			e: e,
		}
		return nil, ErrorSendRequest.ErrorParent(e)
	}

	return rsp, nil
}

func (r *request) _MakeRequest(ctx context.Context) (*http.Request, liberr.Error) {
	var (
		req *http.Request
		err error
	)

	req, err = http.NewRequestWithContext(ctx, r.m, r.u.String(), r.b)

	if err != nil {
		return nil, ErrorCreateRequest.ErrorParent(err)
	}

	if len(r.h) > 0 {
		for k := range r.h {
			req.Header.Set(k, r.h.Get(k))
		}
	}

	q := req.URL.Query()
	for k := range r.p {
		q.Add(k, r.p.Get(k))
	}
	req.URL.RawQuery = q.Encode()

	return req, nil
}

func (r *request) DoParse(ctx context.Context, model interface{}, validStatus ...int) liberr.Error {
	var (
		e error
		b = bytes.NewBuffer(make([]byte, 0))

		err liberr.Error
		rsp *http.Response
	)

	if rsp, err = r.Do(ctx); err != nil {
		return err
	} else if rsp == nil {
		return ErrorResponseInvalid.Error(nil)
	}

	defer func() {
		if !rsp.Close && rsp.Body != nil {
			_ = rsp.Body.Close()
		}
	}()

	if rsp.Body != nil {
		if _, e = io.Copy(b, rsp.Body); e != nil {
			r.e = &requestError{
				c: rsp.StatusCode,
				s: rsp.Status,
				b: b,
				e: e,
			}
			return ErrorResponseLoadBody.ErrorParent(e)
		}
	}

	if !r._IsValidCode(validStatus, rsp.StatusCode) {
		r.e = &requestError{
			c: rsp.StatusCode,
			s: rsp.Status,
			b: b,
			e: nil,
		}
		return ErrorResponseStatus.Error(nil)
	}

	if e = json.Unmarshal(b.Bytes(), model); e != nil {
		r.e = &requestError{
			c: rsp.StatusCode,
			s: rsp.Status,
			b: b,
			e: e,
		}
		return ErrorResponseUnmarshall.ErrorParent(e)
	}

	return nil
}