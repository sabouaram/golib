/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package progress

import (
	"sync/atomic"
	"time"

	liberr "github.com/nabbar/golib/errors"
	libsem "github.com/nabbar/golib/semaphore"
	libmpb "github.com/vbauerster/mpb/v5"
)

type bar struct {
	i *atomic.Value
	s libsem.Sem
	t int64
	b *atomic.Value
	u bool
	w bool
	d bool
}

func newBar(b *libmpb.Bar, s libsem.Sem, total int64, isModeUnic bool) Bar {
	mpbBar := new(atomic.Value)
	mpbBar.Store(b)

	return &bar{
		u: total > 0,
		t: total,
		b: mpbBar,
		s: s,
		w: isModeUnic,
		i: new(atomic.Value),
		d: false,
	}
}

func (b *bar) storeTime(ts time.Time) {
	if b.i == nil {
		b.i = new(atomic.Value)
	}

	b.i.Store(ts)
}

func (b *bar) loadTime() time.Time {
	if b.i == nil {
		b.i = new(atomic.Value)
	}

	if i := b.i.Load(); i == nil {
		return libsem.EmptyTime()
	} else if ts, ok := i.(time.Time); !ok {
		return libsem.EmptyTime()
	} else {
		return ts
	}
}

func (b *bar) storeBar(mpbBar *libmpb.Bar) {
	if b.b == nil {
		b.b = new(atomic.Value)
	}

	b.b.Store(mpbBar)
}

func (b *bar) loadBar() *libmpb.Bar {
	if b.b == nil {
		b.b = new(atomic.Value)
	}

	if i := b.b.Load(); i == nil {
		return nil
	} else if mpbBar, ok := i.(*libmpb.Bar); !ok {
		return nil
	} else {
		return mpbBar
	}
}

func (b bar) GetBarMPB() *libmpb.Bar {
	return b.loadBar()
}

func (b bar) Current() int64 {
	if mpgBar := b.loadBar(); mpgBar == nil {
		return 0
	} else {
		return mpgBar.Current()
	}
}

func (b bar) Completed() bool {
	if mpgBar := b.loadBar(); mpgBar == nil {
		return false
	} else {
		return mpgBar.Completed()
	}
}

func (b *bar) Increment(n int) {
	if n > 0 {
		var mpgBar *libmpb.Bar

		if mpgBar = b.loadBar(); mpgBar == nil {
			panic(ErrorBarNotInitialized.Error(nil))
		}

		mpgBar.IncrBy(n)

		if b.loadTime() == libsem.EmptyTime() {
			b.storeTime(time.Now())
			mpgBar.DecoratorEwmaUpdate(time.Since(b.loadTime()))
		}

		b.storeBar(mpgBar)
	}
}

func (b *bar) Increment64(n int64) {
	if n > 0 {
		var mpgBar *libmpb.Bar

		if mpgBar = b.loadBar(); mpgBar == nil {
			panic(ErrorBarNotInitialized.Error(nil))
		}

		mpgBar.IncrInt64(n)

		if b.loadTime() == libsem.EmptyTime() {
			b.storeTime(time.Now())
			mpgBar.DecoratorEwmaUpdate(time.Since(b.loadTime()))
		}

		b.storeBar(mpgBar)
	}
}

func (b *bar) ResetDefined(current int64) {
	var mpgBar *libmpb.Bar

	if mpgBar = b.loadBar(); mpgBar == nil {
		return
	} else if current >= b.t {
		mpgBar.SetTotal(b.t, true)
		mpgBar.SetRefill(b.t)
	} else {
		mpgBar.SetTotal(b.t, false)
		mpgBar.SetRefill(current)
	}

	b.storeBar(mpgBar)
}

func (b *bar) Reset(total, current int64) {
	b.u = total > 0
	b.t = total
	b.ResetDefined(current)
}

func (b *bar) Done() {
	var mpgBar *libmpb.Bar

	if mpgBar = b.loadBar(); mpgBar == nil {
		return
	}

	mpgBar.SetRefill(b.t)
	mpgBar.SetTotal(b.t, true)
	b.storeBar(mpgBar)
}

func (b *bar) NewWorker() liberr.Error {
	var mpgBar *libmpb.Bar

	if !b.u {
		b.t++
		if mpgBar = b.loadBar(); mpgBar == nil {
			return ErrorBarNotInitialized.Error(nil)
		} else {
			mpgBar.SetTotal(b.t, false)
			b.storeBar(mpgBar)
		}
	}

	if !b.w {
		return b.s.NewWorker()
	}

	return nil
}

func (b *bar) NewWorkerTry() bool {
	if !b.w {
		return b.s.NewWorkerTry()
	}

	return false
}

func (b *bar) DeferWorker() {
	b.Increment(1)
	b.s.DeferWorker()
}

func (b *bar) DeferMain() {
	var mpgBar *libmpb.Bar

	if mpgBar = b.loadBar(); mpgBar == nil {
		return
	} else {
		mpgBar.Abort(b.d)
		b.storeBar(mpgBar)
	}

	if !b.w {
		b.s.DeferMain()
	}
}

func (b *bar) DropOnDefer(flag bool) {
	b.d = flag
}

func (b *bar) WaitAll() liberr.Error {
	if !b.w {
		return b.s.WaitAll()
	}

	return nil
}
