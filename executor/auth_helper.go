// Copyright (C) 2025 JustCoding247. All rights reserved.

package executor

import (
	"fmt"
	"strings"

	"github.com/sqllabs/mysqlaudit/ast"
	"github.com/sqllabs/mysqlaudit/mysql"
	"github.com/sqllabs/mysqlaudit/sessionctx"
	"github.com/sqllabs/mysqlaudit/sessionctx/variable"
	"github.com/sqllabs/mysqlaudit/util/auth"
	"github.com/sqllabs/mysqlaudit/util/sqlexec"
	"github.com/pingcap/errors"
)

const cachingSha2HashLen = 64

func defaultAuthPlugin(vars *variable.SessionVars) string {
	if vars == nil {
		return mysql.AuthCachingSha2Password
	}
	plugin, ok := vars.GetSystemVar("default_authentication_plugin")
	if !ok || plugin == "" {
		return mysql.AuthCachingSha2Password
	}
	return mysql.NormalizeAuthPlugin(plugin)
}

func encodePasswordForAuthOption(vars *variable.SessionVars, opt *ast.AuthOption) (string, string, error) {
	plugin := defaultAuthPlugin(vars)
	if opt == nil {
		return "", plugin, nil
	}
	if len(opt.HashString) > 0 && !opt.ByAuthString {
		hash := opt.HashString
		if detected := inferPluginFromHash(hash); detected != "" {
			plugin = detected
		}
		return hash, plugin, nil
	}
	if opt.ByAuthString {
		hash, err := auth.EncodePasswordByPlugin(plugin, opt.AuthString)
		return hash, plugin, err
	}
	return "", plugin, nil
}

func inferPluginFromHash(hash string) string {
	if len(hash) < 1 || !strings.HasPrefix(hash, "*") {
		return ""
	}
	switch len(hash) - 1 {
	case mysql.PWDHashLen:
		return mysql.AuthNativePassword
	case cachingSha2HashLen:
		return mysql.AuthCachingSha2Password
	default:
		return ""
	}
}

func passwordColumnsForPlugin(plugin, hash string) (passwordCol, authString string) {
	authString = hash
	if plugin == mysql.AuthNativePassword {
		passwordCol = hash
	} else {
		passwordCol = ""
	}
	return
}

func getUserAuthPlugin(ctx sessionctx.Context, user, host string) (string, error) {
	sql := fmt.Sprintf(`SELECT plugin FROM %s.%s WHERE User="%s" AND Host="%s"`, mysql.SystemDB, mysql.UserTable, user, host)
	rows, _, err := ctx.(sqlexec.RestrictedSQLExecutor).ExecRestrictedSQL(ctx, sql)
	if err != nil {
		return "", errors.Trace(err)
	}
	if len(rows) == 0 {
		return defaultAuthPlugin(ctx.GetSessionVars()), nil
	}
	plugin := rows[0].GetString(0)
	return mysql.NormalizeAuthPlugin(plugin), nil
}
