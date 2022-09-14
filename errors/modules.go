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

package errors

const (
	MinPkgArchive     = 100
	MinPkgArtifact    = 200
	MinPkgCertificate = 300
	MinPkgCluster     = 400
	MinPkgConfig      = 500
	MinPkgConsole     = 800
	MinPkgCrypt       = 900
	MinPkgDatabase    = 1000
	MinPkgFTPClient   = 1100
	MinPkgHttpCli     = 1200
	MinPkgHttpServer  = 1300
	MinPkgIOUtils     = 1400
	MinPkgLDAP        = 1500
	MinPkgLogger      = 1600
	MinPkgMail        = 1700
	MinPkgMailer      = 1800
	MinPkgMailPooler  = 1900
	MinPkgNetwork     = 2000
	MinPkgNats        = 2100
	MinPkgNutsDB      = 2200
	MinPkgOAuth       = 2300
	MinPkgAws         = 2400
	MinPkgRequest     = 2500
	MinPkgRouter      = 2600
	MinPkgSemaphore   = 2700
	MinPkgSMTP        = 2800
	MinPkgStatic      = 2900
	MinPkgVersion     = 3000
	MinPkgViper       = 3100

	MinAvailable = 4000

	// MIN_AVAILABLE @Deprecated use MinAvailable constant
	MIN_AVAILABLE = MinAvailable
)
