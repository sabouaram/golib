/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package aws

import (
	"net/http"
	"sync"

	libaws "github.com/nabbar/golib/aws"
	libtls "github.com/nabbar/golib/certificates"
	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
)

type ComponentAwsClient interface {
	cfgtps.Component
	RegisterHTTPClient(fct func() *http.Client)
}

type ComponentAwsAPI interface {
	cfgtps.Component
	RegisterTLS(fct libtls.FctTLSDefault)
}

type ComponentAws interface {
	ComponentAwsClient
	ComponentAwsAPI

	GetAws() (libaws.AWS, error)
	SetAws(a libaws.AWS)
}

func New(drv ConfigDriver) ComponentAws {
	return &componentAws{
		m: sync.RWMutex{},
		x: nil,
		d: drv,
		a: nil,
	}
}

func Register(cfg libcfg.Config, key string, cpt ComponentAws) {
	cfg.ComponentSet(key, cpt)
}

func RegisterNew(cfg libcfg.Config, drv ConfigDriver, key string) {
	cfg.ComponentSet(key, New(drv))
}

func Load(getCpt cfgtps.FuncCptGet, key string) ComponentAws {
	if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(ComponentAws); !ok {
		return nil
	} else {
		return h
	}
}
