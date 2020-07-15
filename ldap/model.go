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

package ldap

import (
	"fmt"
)

type TLSMode uint8

const (
	tlsmode_init TLSMode = iota
	//TLSModeNone no tls connection
	TLSMODE_NONE TLSMode = iota + 1
	//TLSModeTLS strict tls connection
	TLSMODE_TLS
	//TLSModeStartTLS starttls connection (tls into a no tls connection)
	TLSMODE_STARTTLS
)

func (m TLSMode) String() string {
	switch m {
	case TLSMODE_STARTTLS:
		return "starttls"
	case TLSMODE_TLS:
		return "tls"
	case TLSMODE_NONE:
		return "none"
	default:
		return "no defined"
	}
}

func GetDefaultAttributes() []string {
	return []string{"givenName", "mail", "uid", "dn"}
}

type Config struct {
	Uri         string `cloud:"uri" mapstructure:"uri" json:"uri" yaml:"uri" toml:"uri"`
	PortLdap    int    `cloud:"port-ldap" mapstructure:"port-ldap" json:"port-ldap" yaml:"port-ldap" toml:"port-ldap"`
	Portldaps   int    `cloud:"port-ldaps" mapstructure:"port-ldaps" json:"port-ldaps" yaml:"port-ldaps" toml:"port-ldaps"`
	Basedn      string `cloud:"basedn" mapstructure:"basedn" json:"basedn" yaml:"basedn" toml:"basedn"`
	FilterGroup string `cloud:"filter-group" mapstructure:"filter-group" json:"filter-group" yaml:"filter-group" toml:"filter-group"`
	FilterUser  string `cloud:"filter-user" mapstructure:"filter-user" json:"filter-user" yaml:"filter-user" toml:"filter-user"`
}

func NewConfig() *Config {
	return &Config{}
}

func (cnf Config) Clone() *Config {
	return &Config{
		Uri:         cnf.Uri,
		PortLdap:    cnf.PortLdap,
		Portldaps:   cnf.Portldaps,
		Basedn:      cnf.Basedn,
		FilterGroup: cnf.FilterGroup,
		FilterUser:  cnf.FilterUser,
	}
}

func (cnf Config) BaseDN() string {
	return cnf.Basedn
}

func (cnf Config) ServerAddr(withTls bool) string {
	if withTls {
		return fmt.Sprintf("%s:%d", cnf.Uri, cnf.Portldaps)
	}

	return fmt.Sprintf("%s:%d", cnf.Uri, cnf.PortLdap)
}

func (cnf Config) PatternFilterGroup() string {
	return cnf.FilterGroup
}

func (cnf Config) PatternFilterUser() string {
	return cnf.FilterUser
}