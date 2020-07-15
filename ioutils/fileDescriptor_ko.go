// +build windows

/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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
 */

package ioutils

import (
	. "github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/ioutils/maxstdio"
)

func systemFileDescriptor(newValue int) (current int, max int, err Error) {
	rLimit := maxstdio.GetMaxStdio()

	if rLimit < 0 {
		// default windows value
		rLimit = 512
	}

	if newValue > rLimit {
		rLimit = int(maxstdio.SetMaxStdio(newValue))
		return SystemFileDescriptor(0)
	}

	return rLimit, rLimit, nil
	//	return 0, 0, fmt.Errorf("rLimit is nor implemented in current system")
}