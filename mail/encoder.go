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

package mail

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type encoder struct {
	w         *bufio.Writer
	charset   string
	usedChars int
}

// newEncoder returns a new mime header encoder that writes to w. The c
// parameter specifies the name of the character set of the text that will be
// encoded. The u parameter indicates how many characters have been used
// already.
func newEncoder(w io.Writer, c string, u int) *encoder {
	return &encoder{bufio.NewWriter(w), strings.ToUpper(c), u}
}

// encode encodes p using the "Q" encoding and writes it to the underlying
// io.Writer. It limits line length to 75 characters.
func (e *encoder) encode(p []byte) (n int, err error) {
	var output bytes.Buffer
	allPrintable := true

	// some lines we encode end in "
	//maxLineLength := 75 - 1
	maxLineLength := 76

	// prevent header injection
	p = e.secureHeader(p)

	// check to see if we have all printable characters
	for _, c := range p {
		if !e.isVchar(c) && !e.isWSP(c) {
			allPrintable = false
			break
		}
	}

	// all characters are printable. just do line folding
	if allPrintable {
		text := string(p)
		words := strings.Split(text, " ")

		lineBuffer := ""
		firstWord := true

		// split the line where necessary
		for _, word := range words {

			newWord := ""
			if !firstWord {
				newWord += " "
			}
			newWord += word

			// check line length
			if (e.usedChars+len(lineBuffer)+len(newWord) /*+len(" ")+len(word)*/) > maxLineLength && (lineBuffer != "" || e.usedChars != 0) {
				output.WriteString(lineBuffer + "\r\n")

				// first word on newline needs a space in front
				if !firstWord {
					lineBuffer = ""
				} else {
					lineBuffer = " "
				}

				// reset since not on the first line anymore
				e.usedChars = 0
			}
			lineBuffer += newWord /*word*/

			firstWord = false
		}

		output.WriteString(lineBuffer)

	} else {
		firstLine := true

		// A single encoded word can not be longer than 75 characters
		if e.usedChars == 0 {
			maxLineLength = 75
		}

		wordBegin := "=?" + e.charset + "?Q?"
		wordEnd := "?="

		lineBuffer := wordBegin

		for i := 0; i < len(p); {
			// encode the character
			encodedChar, runeLength := e.encodeChar(p, i)

			// Check line length
			if len(lineBuffer)+e.usedChars+len(encodedChar) > (maxLineLength - len(wordEnd)) {
				output.WriteString(lineBuffer + wordEnd + "\r\n")
				lineBuffer = " " + wordBegin
				firstLine = false
			}

			lineBuffer += encodedChar

			i = i + runeLength

			// reset since not on the first line anymore
			if !firstLine {
				e.usedChars = 0
				maxLineLength = 76
			}
		}

		output.WriteString(lineBuffer + wordEnd)
	}

	_, _ = e.w.Write(output.Bytes())
	_ = e.w.Flush()
	n = output.Len()

	return n, nil
}

// encode takes a string and position in that string and encodes one utf-8
// character. It then returns the encoded string and number of runes in the
// character.
func (e *encoder) encodeChar(text []byte, i int) (encodedString string, runeLength int) {
	started := false

	for ; i < len(text) && (!utf8.RuneStart(text[i]) || !started); i++ {
		switch c := text[i]; {
		case c == ' ':
			encodedString += "_"
		case e.isVchar(c) && c != '=' && c != '?' && c != '_':
			encodedString += string(c)
		default:
			encodedString += fmt.Sprintf("=%02X", c)
		}

		runeLength++

		started = true
	}

	return
}

// secureHeader removes all unnecessary values to prevent
// header injection
func (e *encoder) secureHeader(text []byte) []byte {
	secureValue := strings.TrimSpace(string(text))
	secureValue = strings.Replace(secureValue, "\r", "", -1)
	secureValue = strings.Replace(secureValue, "\n", "", -1)
	secureValue = strings.Replace(secureValue, "\t", "", -1)

	return []byte(secureValue)
}

// isVchar returns true if c is an RFC 5322 VCHAR character.
func (e *encoder) isVchar(c byte) bool {
	// Visible (printing) characters.
	return '!' <= c && c <= '~'
}

// isWSP returns true if c is a WSP (white space).
// WSP is a space or horizontal tab (RFC5234 Appendix B).
func (e *encoder) isWSP(c byte) bool {
	return c == ' ' || c == '\t'
}