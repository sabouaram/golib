/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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

package udp_test

import (
	"context"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/*
	Using https://onsi.github.io/ginkgo/
	Running with $> ginkgo -cover .
*/

var (
	ctx context.Context
	cnl context.CancelFunc
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// TestGolibEncodingMuxHelper tests the Golib Mux Encoding Helper function.
func TestGolibSocketServerUdpHelper(t *testing.T) {
	ctx, cnl = context.WithCancel(context.Background())
	defer cnl()

	time.Sleep(500 * time.Millisecond)     // Adding delay for better testing synchronization
	RegisterFailHandler(Fail)              // Registering fail handler for better test failure reporting
	RunSpecs(t, "Socket Server UDP Suite") // Running the test suite for Encoding Mux Helper
}

var _ = BeforeSuite(func() {
})

var _ = AfterSuite(func() {
})
