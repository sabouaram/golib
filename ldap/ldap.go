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
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"github.com/go-ldap/ldap/v3"
	libcrt "github.com/nabbar/golib/certificates"
	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
)

//HelperLDAP struct use to manage connection to server and request it.
type HelperLDAP struct {
	Attributes []string
	conn       *ldap.Conn
	config     *Config
	tlsConfig  *tls.Config
	tlsMode    TLSMode
	bindDN     string
	bindPass   string
	ctx        context.Context
}

//NewLDAP build a new LDAP helper based on config struct given.
func NewLDAP(ctx context.Context, cnf *Config, attributes []string) (*HelperLDAP, liberr.Error) {
	if cnf == nil {
		return nil, ErrorEmptyParams.Error(nil)
	}

	return &HelperLDAP{
		Attributes: attributes,
		tlsConfig:  libcrt.GetTLSConfig(cnf.Uri),
		tlsMode:    tlsmode_init,
		config:     cnf.Clone(),
		ctx:        libctx.IsolateParent(ctx),
	}, nil
}

//SetCredentials used to defined the BindDN and password for connection.
func (lc *HelperLDAP) SetCredentials(user, pass string) {
	lc.bindDN = user
	lc.bindPass = pass
}

//ForceTLSMode used to force tls mode and defined tls condition.
func (lc *HelperLDAP) ForceTLSMode(tlsMode TLSMode, tlsConfig *tls.Config) {
	if tlsConfig != nil {
		lc.tlsConfig = tlsConfig
	} else {
		//nolint #nosec
		/* #nosec */
		lc.tlsConfig = &tls.Config{}
	}

	switch tlsMode {
	case TLSMODE_TLS:
		lc.tlsMode = TLSMODE_TLS
	case TLSMODE_STARTTLS:
		lc.tlsMode = TLSMODE_STARTTLS
	case TLSMODE_NONE:
		lc.tlsConfig = nil
		lc.tlsMode = TLSMODE_NONE
	case tlsmode_init:
		lc.tlsMode = tlsmode_init
	}
}

func (lc *HelperLDAP) dialTLS() (*ldap.Conn, liberr.Error) {
	d := net.Dialer{}

	c, err := d.DialContext(lc.ctx, "tcp", lc.config.ServerAddr(true))

	if err != nil {
		if c != nil {
			_ = c.Close()
		}

		return nil, ErrorLDAPServerTLS.ErrorParent(err)
	}

	c = tls.Client(c, lc.tlsConfig)

	if c == nil {
		return nil, ErrorLDAPServerTLS.ErrorParent(ErrorLDAPServerConnection.Error(nil))
	}

	l := ldap.NewConn(c, true)
	if l == nil {
		return nil, ErrorLDAPServerTLS.ErrorParent(ErrorLDAPServerConnection.Error(nil))
	}

	l.Start()

	if l.IsClosing() {
		return nil, ErrorLDAPServerTLS.ErrorParent(ErrorLDAPServerDialClosing.Error(nil))
	}

	if _, tlsOk := l.TLSConnectionState(); !tlsOk {
		return nil, ErrorLDAPServerTLS.ErrorParent(nil)
	}

	return l, nil
}

func (lc *HelperLDAP) dial() (*ldap.Conn, liberr.Error) {
	d := net.Dialer{}

	c, err := d.DialContext(lc.ctx, "tcp", lc.config.ServerAddr(true))

	if err != nil {
		if c != nil {
			_ = c.Close()
		}

		return nil, ErrorLDAPServerDial.ErrorParent(err)
	}

	l := ldap.NewConn(c, false)
	if l == nil {
		return nil, ErrorLDAPServerDial.ErrorParent(ErrorLDAPServerConnection.Error(nil))
	}

	l.Start()

	if l.IsClosing() {
		return nil, ErrorLDAPServerDial.ErrorParent(ErrorLDAPServerDialClosing.Error(nil))
	}

	return l, nil
}

func (lc *HelperLDAP) starttls(l *ldap.Conn) liberr.Error {
	err := l.StartTLS(lc.tlsConfig)

	if err != nil {
		return ErrorLDAPServerStartTLS.ErrorParent(err)
	}

	if _, tlsOk := l.TLSConnectionState(); !tlsOk {
		return ErrorLDAPServerStartTLS.ErrorParent(nil)
	}

	return nil
}

func (lc *HelperLDAP) tryConnect() (TLSMode, liberr.Error) {
	if lc == nil {
		return TLSMODE_NONE, ErrorEmptyParams.Error(nil)
	}

	var (
		l   *ldap.Conn
		err liberr.Error
	)

	defer func() {
		if l != nil {
			l.Close()
		}
	}()

	if lc.config.Portldaps != 0 {
		l, err = lc.dialTLS()

		liblog.DebugLevel.LogErrorCtxf(liblog.DebugLevel, "connecting ldap with tls mode '%s'", err, TLSMODE_TLS.String())

		if err == nil {
			return TLSMODE_TLS, nil
		}
	}

	if lc.config.PortLdap == 0 {
		return tlsmode_init, ErrorLDAPServerConfig.Error(nil)
	}

	l, err = lc.dial()
	liblog.DebugLevel.LogErrorCtxf(liblog.DebugLevel, "connecting ldap with tls mode '%s'", err, TLSMODE_NONE.String())

	if err != nil {
		return tlsmode_init, err
	}

	err = lc.starttls(l)
	liblog.DebugLevel.LogErrorCtxf(liblog.DebugLevel, "connecting ldap with tls mode '%s'", err, TLSMODE_STARTTLS.String())

	if err == nil {
		return TLSMODE_STARTTLS, nil
	}

	return TLSMODE_NONE, nil
}

func (lc *HelperLDAP) connect() liberr.Error {
	if lc == nil || lc.ctx == nil {
		return ErrorLDAPContext.Error(ErrorEmptyParams.Error(nil))
	}

	if err := lc.ctx.Err(); err != nil {
		return ErrorLDAPContext.ErrorParent(err)
	}

	if lc.conn == nil {
		var (
			l   *ldap.Conn
			err liberr.Error
		)

		if lc.tlsMode == tlsmode_init {
			m, e := lc.tryConnect()

			if e != nil {
				return e
			}

			lc.tlsMode = m
		}

		if lc.tlsMode == TLSMODE_TLS {
			l, err = lc.dialTLS()
			if err != nil {
				if l != nil {
					l.Close()
				}
				return err
			}
		}

		if lc.tlsMode == TLSMODE_NONE || lc.tlsMode == TLSMODE_STARTTLS {
			l, err = lc.dial()
			if err != nil {
				if l != nil {
					l.Close()
				}
				return err
			}
		}

		if lc.tlsMode == TLSMODE_STARTTLS {
			err = lc.starttls(l)
			if err != nil {
				if l != nil {
					l.Close()
				}
				return err
			}
		}

		liblog.DebugLevel.Logf("ldap connected with tls mode '%s'", lc.tlsMode.String())
		lc.conn = l
	}

	return nil
}

//Check used to check if connection success (without any bind).
func (lc *HelperLDAP) Check() liberr.Error {
	if lc == nil {
		return ErrorEmptyParams.Error(nil)
	}

	if lc.conn == nil {
		defer func() {
			if lc.conn != nil {
				lc.conn.Close()
				lc.conn = nil
			}
		}()
	}

	if err := lc.connect(); err != nil {
		lc.Close()
		return err
	}

	return nil
}

//Close used to close connection object.
func (lc *HelperLDAP) Close() {
	if lc == nil {
		return
	}

	if lc.conn != nil {
		lc.conn.Close()
		lc.conn = nil
	}
}

//AuthUser used to test bind given user uid and password.
func (lc *HelperLDAP) AuthUser(username, password string) liberr.Error {
	if lc == nil {
		return ErrorEmptyParams.Error(nil)
	}

	if err := lc.connect(); err != nil {
		return err
	}

	if username == "" || password == "" {
		return ErrorEmptyParams.Error(nil)
	}

	err := lc.conn.Bind(username, password)

	return ErrorLDAPBind.Iferror(err)
}

//Connect used to connect and bind to server.
func (lc *HelperLDAP) Connect() liberr.Error {
	if lc == nil {
		return ErrorEmptyParams.Error(nil)
	}

	if err := lc.AuthUser(lc.bindDN, lc.bindPass); err != nil {
		return err
	}

	liblog.DebugLevel.Logf("Bind success on LDAP server %s with tls mode '%s'", lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String())
	return nil
}

func (lc *HelperLDAP) runSearch(filter string, attributes []string) (*ldap.SearchResult, liberr.Error) {
	var (
		err error
		src *ldap.SearchResult
	)

	if e := lc.Connect(); e != nil {
		return nil, e
	}

	defer lc.Close()

	searchRequest := ldap.NewSearchRequest(
		lc.config.Basedn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		100, 0, false,
		filter,
		attributes,
		nil,
	)

	if src, err = lc.conn.Search(searchRequest); err != nil {
		return nil, ErrorLDAPSearch.ErrorParent(err)
	}

	liblog.DebugLevel.Logf("Search success on server '%s' with tls mode '%s', with filter [%s] and attribute %v", lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), filter, attributes)
	return src, nil
}

func (lc *HelperLDAP) getUserName(username string) (string, liberr.Error) {
	username = strings.TrimSpace(username)
	if username == "" {
		if usr := lc.ParseEntries(lc.bindDN); len(usr) == 0 {
			return "", ErrorLDAPInvalidUID.Error(ErrorLDAPInvalidDN.Error(nil))
		} else if _, ok := usr["uid"]; !ok {
			return "", ErrorLDAPInvalidUID.Error(ErrorLDAPAttributeNotFound.Error(nil))
		} else if len(usr["uid"]) < 1 {
			return "", ErrorLDAPInvalidUID.Error(ErrorLDAPAttributeEmpty.Error(nil))
		} else {
			username = usr["uid"][0]
		}

		username = strings.TrimSpace(username)
	}

	if username == "" {
		return "", ErrorLDAPInvalidUID.Error(ErrorLDAPAttributeEmpty.Error(nil))
	}

	return username, nil
}

//UserInfo used to retrieve the information of a given username.
func (lc *HelperLDAP) UserInfo(username string) (map[string]string, liberr.Error) {
	var (
		err     liberr.Error
		src     *ldap.SearchResult
		userRes map[string]string
	)

	if username, err = lc.getUserName(username); err != nil {
		return nil, err
	}

	userRes = make(map[string]string)
	attributes := append(lc.Attributes, "cn")

	src, err = lc.runSearch(fmt.Sprintf(lc.config.FilterUser, userFieldUid, username), attributes)

	if err != nil {
		return userRes, err
	}

	if len(src.Entries) != 1 {
		if len(src.Entries) > 1 {
			return userRes, ErrorLDAPUserNotUniq.Error(nil)
		} else {
			return userRes, ErrorLDAPUserNotFound.Error(nil)
		}
	}

	for _, attr := range attributes {
		userRes[attr] = src.Entries[0].GetAttributeValue(attr)
	}

	if _, ok := userRes["DN"]; !ok {
		userRes["DN"] = src.Entries[0].DN
	}

	liblog.DebugLevel.Logf("Map info retrieve in ldap server '%s' with tls mode '%s' about user [%s] : %v", lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), username, userRes)
	return userRes, nil
}

//UserInfo used to retrieve the information of a given username.
func (lc *HelperLDAP) UserInfoByField(username string, fieldOfUnicValue string) (map[string]string, liberr.Error) {
	var (
		err     liberr.Error
		src     *ldap.SearchResult
		userRes map[string]string
	)

	if username, err = lc.getUserName(username); err != nil {
		return nil, err
	}

	userRes = make(map[string]string)
	attributes := append(lc.Attributes, "cn")

	src, err = lc.runSearch(fmt.Sprintf(lc.config.FilterUser, fieldOfUnicValue, username), attributes)

	if err != nil {
		return userRes, err
	}

	if len(src.Entries) != 1 {
		if len(src.Entries) > 1 {
			return userRes, ErrorLDAPUserNotUniq.Error(nil)
		} else {
			return userRes, ErrorLDAPUserNotFound.Error(nil)
		}
	}

	for _, attr := range attributes {
		userRes[attr] = src.Entries[0].GetAttributeValue(attr)
	}

	if _, ok := userRes["DN"]; !ok {
		userRes["DN"] = src.Entries[0].DN
	}

	liblog.DebugLevel.Logf("Map info retrieve in ldap server '%s' with tls mode '%s' about user [%s] : %v", lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), username, userRes)
	return userRes, nil
}

//GroupInfo used to retrieve the information of a given group cn.
func (lc *HelperLDAP) GroupInfo(groupname string) (map[string]interface{}, liberr.Error) {
	var (
		err     liberr.Error
		src     *ldap.SearchResult
		grpInfo map[string]interface{}
	)

	src, err = lc.runSearch(fmt.Sprintf(lc.config.FilterGroup, groupFieldCN, groupname), []string{})
	if err != nil {
		return grpInfo, err
	}

	if len(src.Entries) == 0 {
		return nil, ErrorLDAPGroupNotFound.Error(nil)
	}

	grpInfo = make(map[string]interface{}, len(src.Entries[0].Attributes))
	for _, entry := range src.Entries {
		for _, entryAttribute := range entry.Attributes {
			grpInfo[entryAttribute.Name] = entryAttribute.Values
		}
	}

	liblog.DebugLevel.Logf("Info for group [%s] find on server '%s' with tls mode '%s' : %v", groupname, lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), grpInfo)
	return grpInfo, nil
}

//GroupInfo used to retrieve the information of a given group cn.
func (lc *HelperLDAP) GroupInfoByField(groupname string, fieldForUnicValue string) (map[string]interface{}, liberr.Error) {
	var (
		err     liberr.Error
		src     *ldap.SearchResult
		grpInfo map[string]interface{}
	)

	src, err = lc.runSearch(fmt.Sprintf(lc.config.FilterGroup, fieldForUnicValue, groupname), []string{})
	if err != nil {
		return grpInfo, err
	}

	if len(src.Entries) == 0 {
		return nil, ErrorLDAPGroupNotFound.Error(nil)
	}

	grpInfo = make(map[string]interface{}, len(src.Entries[0].Attributes))
	for _, entry := range src.Entries {
		for _, entryAttribute := range entry.Attributes {
			grpInfo[entryAttribute.Name] = entryAttribute.Values
		}
	}

	liblog.DebugLevel.Logf("Info for group [%s] find on server '%s' with tls mode '%s' : %v", groupname, lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), grpInfo)
	return grpInfo, nil
}

//UserMemberOf returns the group list of a given user.
func (lc *HelperLDAP) UserMemberOf(username string) ([]string, liberr.Error) {
	var (
		err liberr.Error
		src *ldap.SearchResult
		grp []string
	)

	if username, err = lc.getUserName(username); err != nil {
		return nil, err
	}

	grp = make([]string, 0)

	src, err = lc.runSearch(fmt.Sprintf(lc.config.FilterUser, userFieldUid, username), []string{"memberOf"})
	if err != nil {
		return grp, err
	}

	for _, entry := range src.Entries {
		for _, mmb := range entry.GetAttributeValues("memberOf") {
			liblog.DebugLevel.Logf("Group find for uid '%s' on server '%s' with tls mode '%s' : %v", username, lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), mmb)
			mmo := lc.ParseEntries(mmb)
			grp = append(grp, mmo["cn"]...)
		}
	}

	liblog.DebugLevel.Logf("Groups find for uid '%s' on server '%s' with tls mode '%s' : %v", username, lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), grp)
	return grp, nil
}

//UserIsInGroup used to check if a given username is a group member of a list of reference group name.
func (lc *HelperLDAP) UserIsInGroup(username string, groupname []string) (bool, liberr.Error) {
	var (
		err     liberr.Error
		grpMmbr []string
	)

	if username, err = lc.getUserName(username); err != nil {
		return false, err
	} else if grpMmbr, err = lc.UserMemberOf(username); err != nil {
		return false, err
	}

	for _, grpSrch := range groupname {
		for _, grpItem := range grpMmbr {
			if strings.EqualFold(grpSrch, grpItem) {
				return true, nil
			}
		}
	}

	return false, nil
}

//UsersOfGroup used to retrieve the member list of a given group name.
func (lc *HelperLDAP) UsersOfGroup(groupname string) ([]string, liberr.Error) {
	var (
		err liberr.Error
		src *ldap.SearchResult
		grp []string
	)

	grp = make([]string, 0)

	src, err = lc.runSearch(fmt.Sprintf(lc.config.FilterGroup, groupFieldCN, groupname), []string{"member"})
	if err != nil {
		return grp, err
	}

	for _, entry := range src.Entries {
		for _, mmb := range entry.GetAttributeValues("member") {
			member := lc.ParseEntries(mmb)
			grp = append(grp, member["uid"]...)
		}
	}

	liblog.DebugLevel.Logf("Member of groups [%s] find on server '%s' with tls mode '%s' : %v", groupname, lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), grp)
	return grp, nil
}

//ParseEntries used to clean attributes of an object class.
func (lc HelperLDAP) ParseEntries(entry string) map[string][]string {
	var listEntries = make(map[string][]string)

	for _, ent := range strings.Split(entry, ",") {
		key := strings.SplitN(ent, "=", 2)

		if len(key) != 2 || len(key[0]) < 1 || len(key[1]) < 1 {
			continue
		}

		key[0] = strings.TrimSpace(key[0])
		key[1] = strings.TrimSpace(key[1])

		if _, ok := listEntries[key[0]]; !ok {
			listEntries[key[0]] = []string{}
		}

		listEntries[key[0]] = append(listEntries[key[0]], key[1])
	}

	return listEntries
}
