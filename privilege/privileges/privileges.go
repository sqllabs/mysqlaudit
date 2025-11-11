// Copyright 2015 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package privileges

import (
	"strings"

	"github.com/sqllabs/mysqlaudit/mysql"
	"github.com/sqllabs/mysqlaudit/privilege"
	"github.com/sqllabs/mysqlaudit/sessionctx"
	"github.com/sqllabs/mysqlaudit/terror"
	"github.com/sqllabs/mysqlaudit/types"
	"github.com/sqllabs/mysqlaudit/util/auth"
	log "github.com/sirupsen/logrus"
)

// SkipWithGrant causes the server to start without using the privilege system at all.
var SkipWithGrant = false

// privilege error codes.
const (
	codeInvalidPrivilegeType  terror.ErrCode = 1
	codeInvalidUserNameFormat                = 2
)

var (
	errInvalidPrivilegeType  = terror.ClassPrivilege.New(codeInvalidPrivilegeType, "unknown privilege type")
	errInvalidUserNameFormat = terror.ClassPrivilege.New(codeInvalidUserNameFormat, "wrong username format")
)

var _ privilege.Manager = (*UserPrivileges)(nil)

// UserPrivileges implements privilege.Manager interface.
// This is used to check privilege for the current user.
type UserPrivileges struct {
	user string
	host string
	*Handle
}

// RequestVerification implements the Manager interface.
func (p *UserPrivileges) RequestVerification(db, table, column string, priv mysql.PrivilegeType) bool {
	if SkipWithGrant {
		return true
	}

	if p.user == "" && p.host == "" {
		return true
	}

	// Skip check for INFORMATION_SCHEMA database.
	// See https://dev.mysql.com/doc/refman/5.7/en/information-schema.html
	if strings.EqualFold(db, "INFORMATION_SCHEMA") {
		return true
	}

	mysqlPriv := p.Handle.Get()
	return mysqlPriv.RequestVerification(p.user, p.host, db, table, column, priv)
}

// ConnectionVerification implements the Manager interface.
func (p *UserPrivileges) ConnectionVerification(user, host string, authentication, salt []byte, clientPlugin string) (u string, h string, plugin string, success bool, err error) {

	if SkipWithGrant {
		p.user = user
		p.host = host
		success = true
		return u, h, mysql.AuthNativePassword, success, nil
	}

	mysqlPriv := p.Handle.Get()
	record := mysqlPriv.connectionVerification(user, host)
	if record == nil {
		log.Errorf("Get user privilege record fail: user %v, host %v", user, host)
		return "", "", "", false, nil
	}

	u = record.User
	h = record.Host
	plugin = record.getAuthPlugin()
	if clientPlugin != "" && plugin != "" && mysql.NormalizeAuthPlugin(clientPlugin) != plugin {
		log.Infof("plugin mismatch user=%s client=%s expected=%s", user, clientPlugin, plugin)
		return u, h, plugin, false, nil
	}

	pwd := record.getAuthString()
	log.Infof("auth plugin=%s authLen=%d storedLen=%d user=%s", plugin, len(authentication), len(pwd), user)

	// empty password
	if len(pwd) == 0 && len(authentication) == 0 {
		p.user = user
		p.host = host
		success = true
		return u, h, plugin, success, nil
	}

	if len(pwd) == 0 || len(authentication) == 0 {
		return u, h, plugin, false, nil
	}

	ok, err := auth.VerifyPassword(plugin, pwd, authentication, salt)
	if err != nil {
		log.Errorf("Verify password error %v", err)
		return u, h, plugin, false, err
	}

	if !ok {
		return u, h, plugin, false, nil
	}

	p.user = user
	p.host = host
	success = true
	return u, h, plugin, success, nil
}

// DBIsVisible implements the Manager interface.
func (p *UserPrivileges) DBIsVisible(db string) bool {
	if SkipWithGrant {
		return true
	}
	mysqlPriv := p.Handle.Get()
	return mysqlPriv.DBIsVisible(p.user, p.host, db)
}

// UserPrivilegesTable implements the Manager interface.
func (p *UserPrivileges) UserPrivilegesTable() [][]types.Datum {
	mysqlPriv := p.Handle.Get()
	return mysqlPriv.UserPrivilegesTable()
}

// ShowGrants implements privilege.Manager ShowGrants interface.
func (p *UserPrivileges) ShowGrants(ctx sessionctx.Context, user *auth.UserIdentity) ([]string, error) {
	mysqlPrivilege := p.Handle.Get()
	return mysqlPrivilege.showGrants(user.Username, user.Hostname), nil
}
