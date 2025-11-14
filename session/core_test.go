package session_test

import (
	"fmt"
	"testing"

	. "github.com/pingcap/check"
	"github.com/sqllabs/mysqlaudit/config"
	"github.com/sqllabs/mysqlaudit/session"
	"golang.org/x/net/context"
)

var _ = Suite(&testInceptionSuite{})

type testInceptionSuite struct{}

func TestInception(t *testing.T) {
	TestingT(t)
}

func (s *testInceptionSuite) SetUpSuite(c *C) {
	inc := &config.GetGlobalConfig().Inc

	inc.BackupHost = "127.0.0.1"
	inc.BackupPort = 3306
	inc.BackupUser = "test"
	inc.BackupPassword = "test"

	inc.Lang = "en-US"
	inc.EnableFingerprint = true
	inc.SqlSafeUpdates = 0
	inc.EnableDropTable = true
}

func (s *testInceptionSuite) TestCheck(c *C) {
	core := session.NewInception()
	opts := loadTestSourceOptions()
	core.LoadOptions(opts)
	sql := wrapSQLWithDB(`drop table if exists t1;
	create table t1(id int primary key);
	insert into t1 values(1);`)
	result, err := core.Audit(context.Background(), sql)
	c.Assert(err, IsNil)

	for _, row := range result {
		// fmt.Println(fmt.Sprintf("%#v", row))
		if row.ErrLevel == 2 {
			fmt.Println(fmt.Sprintf("sql: %v, err: %v", row.Sql, row.ErrorMessage))
		} else {
			fmt.Println(fmt.Sprintf("[%v] sql: %v", session.StatusList[row.StageStatus], row.Sql))
		}
	}
}

func (s *testInceptionSuite) TestExecute(c *C) {
	core := session.NewInception()
	opts := loadTestSourceOptions()
	core.LoadOptions(opts)
	sql := wrapSQLWithDB(`drop table if exists t1;
	create table t1(id int primary key);
	insert into t1 values(1);`)
	result, err := core.RunExecute(context.Background(), sql)
	c.Assert(err, IsNil)

	for _, row := range result {
		// fmt.Println(fmt.Sprintf("%#v", row))
		if row.ErrLevel == 2 {
			fmt.Println(fmt.Sprintf("sql: %v, err: %v", row.Sql, row.ErrorMessage))
		} else {
			fmt.Println(fmt.Sprintf("[%v] sql: %v", session.StatusList[row.StageStatus], row.Sql))
		}
	}
}

func (s *testInceptionSuite) TestBackup(c *C) {
	core := session.NewInception()
	opts := loadTestSourceOptions()
	opts.Backup = true
	core.LoadOptions(opts)
	sql := wrapSQLWithDB(`drop table if exists t1;
	create table t1(id int primary key);
	insert into t1 values(1);`)
	result, err := core.RunExecute(context.Background(), sql)
	c.Assert(err, IsNil)

	for _, row := range result {
		// fmt.Println(fmt.Sprintf("%#v", row))
		if row.ErrLevel == 2 {
			fmt.Println(fmt.Sprintf("sql: %v, err: %v", row.Sql, row.ErrorMessage))
		} else {
			fmt.Println(fmt.Sprintf("[%v] sql: %v", session.StatusList[row.StageStatus], row.Sql))
		}
	}
}

func (s *testInceptionSuite) TestDropTable(c *C) {
	core := session.NewInception()
	opts := loadTestSourceOptions()
	opts.Backup = true
	core.LoadOptions(opts)
	sql := wrapSQLWithDB(`drop table if exists t000001;
	create table t000001(id int);`)
	result, err := core.RunExecute(context.Background(), sql)
	c.Assert(err, IsNil)

	for _, row := range result {
		// fmt.Println(fmt.Sprintf("%#v", row))
		if row.ErrLevel == 2 {
			fmt.Println(fmt.Sprintf("sql: %v, err: %v", row.Sql, row.ErrorMessage))
		} else {
			fmt.Println(fmt.Sprintf("[%v] sql: %v", session.StatusList[row.StageStatus], row.Sql))
		}
	}
}

func loadTestSourceOptions() session.SourceOptions {
	return session.SourceOptions{
		Host:     getTestHost(),
		Port:     int(getTestPort()),
		User:     getTestUser(),
		Password: getTestPassword(),
	}
}

func wrapSQLWithDB(sql string) string {
	return fmt.Sprintf("use %s;\n%s", getTestDBName(), sql)
}
