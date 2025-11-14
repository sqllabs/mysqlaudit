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

package session_test

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/pingcap/check"
	"github.com/sqllabs/mysqlaudit/config"
	"github.com/sqllabs/mysqlaudit/util/testkit"
)

var _ = Suite(&testSessionMaskingSuite{})

func TestMasking(t *testing.T) {
	TestingT(t)
}

type testSessionMaskingSuite struct {
	testCommon
}

func (s *testSessionMaskingSuite) SetUpSuite(c *C) {

	s.initSetUp(c)

	config.GetGlobalConfig().Inc.Lang = "en-US"
	config.GetGlobalConfig().Inc.EnableFingerprint = true
	config.GetGlobalConfig().Inc.SqlSafeUpdates = 0

	if s.tk == nil {
		s.tk = testkit.NewTestKitWithInit(c, s.store)
	}
}

func (s *testSessionMaskingSuite) TearDownSuite(c *C) {
	s.tearDownSuite(c)
}

func (s *testSessionMaskingSuite) makeSQL(sql string) *testkit.Result {
	comment := s.buildOptionComment("--masking=1", "--enable-ignore-warnings")
	a := fmt.Sprintf(`%s
inception_magic_start;
%s
%s;
inception_magic_commit;`, comment, s.useDB, "%s")
	return s.tk.MustQueryInc(fmt.Sprintf(a, sql))
}

type maskingField struct {
	Index  int    `json:"index"`
	Field  string `json:"field"`
	Type   string `json:"type"`
	Table  string `json:"table"`
	Schema string `json:"schema"`
	Alias  string `json:"alias"`
}

func expectMaskingFields(fields ...maskingField) string {
	bytes, err := json.Marshal(fields)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func (s *testSessionMaskingSuite) TestInsert(c *C) {

	res := s.makeSQL("insert into t1 values(1);")
	row := res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "2", Commentf("%v", row))
	c.Assert(row[4], Equals, "not support", Commentf("%v", row))

	res = s.makeSQL("insert into t1 values;")

	if len(res.Rows()) > 0 {
		row = res.Rows()[int(s.tk.Se.AffectedRows())-1]
		c.Assert(row[2], Equals, "2", Commentf("%v", row))
		c.Assert(row[4], Equals, "line 1 column 21 near \"\" ", Commentf("%v", row))
	} else {
		c.Assert(len(res.Rows()), Greater, 0, Commentf("%v", res))
	}
}

func (s *testSessionMaskingSuite) TestQuery(c *C) {

	s.mustRunExec(c, `drop table if exists t1,t2;
	create table t1(id int primary key,c1 int);
	create table t2(id int primary key,c1 int,c2 int);
	insert into t1 values(1,1),(2,2);`)

	res := s.makeSQL(`select * from t1;`)
	row := res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "0", Commentf("%v", row))
	c.Assert(row[3], Equals, expectMaskingFields(
		maskingField{Index: 0, Field: "id", Type: "int", Table: "t1", Schema: "test_inc", Alias: "id"},
		maskingField{Index: 1, Field: "c1", Type: "int", Table: "t1", Schema: "test_inc", Alias: "c1"},
	), Commentf("%v", row))

	res = s.makeSQL(`select a.* from t1 a;`)
	row = res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "0", Commentf("%v", row))
	c.Assert(row[3], Equals, expectMaskingFields(
		maskingField{Index: 0, Field: "id", Type: "int", Table: "t1", Schema: "test_inc", Alias: "id"},
		maskingField{Index: 1, Field: "c1", Type: "int", Table: "t1", Schema: "test_inc", Alias: "c1"},
	), Commentf("%v", row))

	res = s.makeSQL(`select * from t1 union select id,c2 from t2;`)
	row = res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "0", Commentf("%v", row))
	c.Assert(row[3], Equals, expectMaskingFields(
		maskingField{Index: 0, Field: "id", Type: "int", Table: "t1", Schema: "test_inc", Alias: "id"},
		maskingField{Index: 1, Field: "c1", Type: "int", Table: "t1", Schema: "test_inc", Alias: "c1"},
		maskingField{Index: 2, Field: "id", Type: "int", Table: "t2", Schema: "test_inc", Alias: "id"},
		maskingField{Index: 3, Field: "c2", Type: "int", Table: "t2", Schema: "test_inc", Alias: "c2"},
	), Commentf("%v", row))

	res = s.makeSQL(`select ifnull(c1,c2) from t1 inner join t2 on t1.id=t2.id`)
	row = res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "0", Commentf("%v", row))
	c.Assert(row[3], Equals, expectMaskingFields(
		maskingField{Index: 0, Field: "c1", Type: "int", Table: "t1", Schema: "test_inc", Alias: "ifnull(c1,c2)"},
		maskingField{Index: 0, Field: "c2", Type: "int", Table: "t2", Schema: "test_inc", Alias: "ifnull(c1,c2)"},
	), Commentf("%v", row))

	res = s.makeSQL(`select a.c1_alias,a.c3_alias from (select *,c1 as c1_alias,concat(id,c2) as c3_alias from t2) a;`)
	row = res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "0", Commentf("%v", row))
	c.Assert(row[3], Equals, expectMaskingFields(
		maskingField{Index: 0, Field: "c1", Type: "int", Table: "t2", Schema: "test_inc", Alias: "c1_alias"},
		maskingField{Index: 1, Field: "id", Type: "int", Table: "t2", Schema: "test_inc", Alias: "c3_alias"},
		maskingField{Index: 1, Field: "c2", Type: "int", Table: "t2", Schema: "test_inc", Alias: "c3_alias"},
	), Commentf("%v", row))

	res = s.makeSQL(`select a.*,concat(a.id,a.c1) as c4 from (select * from t2) a;`)
	row = res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "0", Commentf("%v", row))
	c.Assert(row[3], Equals, expectMaskingFields(
		maskingField{Index: 0, Field: "id", Type: "int", Table: "t2", Schema: "test_inc", Alias: "id"},
		maskingField{Index: 1, Field: "c1", Type: "int", Table: "t2", Schema: "test_inc", Alias: "c1"},
		maskingField{Index: 2, Field: "c2", Type: "int", Table: "t2", Schema: "test_inc", Alias: "c2"},
		maskingField{Index: 3, Field: "id", Type: "int", Table: "t2", Schema: "test_inc", Alias: "c4"},
		maskingField{Index: 3, Field: "c1", Type: "int", Table: "t2", Schema: "test_inc", Alias: "c4"},
	), Commentf("%v", row))

	res = s.makeSQL(`select a1.*,a1.id,a2.* from t1 a1 inner join t2 a2 on a1.id=a2.id`)
	row = res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "0", Commentf("%v", row))
	c.Assert(row[3], Equals, expectMaskingFields(
		maskingField{Index: 0, Field: "id", Type: "int", Table: "t1", Schema: "test_inc", Alias: "id"},
		maskingField{Index: 1, Field: "c1", Type: "int", Table: "t1", Schema: "test_inc", Alias: "c1"},
		maskingField{Index: 2, Field: "id", Type: "int", Table: "t1", Schema: "test_inc", Alias: "id"},
		maskingField{Index: 3, Field: "id", Type: "int", Table: "t2", Schema: "test_inc", Alias: "id"},
		maskingField{Index: 4, Field: "c1", Type: "int", Table: "t2", Schema: "test_inc", Alias: "c1"},
		maskingField{Index: 5, Field: "c2", Type: "int", Table: "t2", Schema: "test_inc", Alias: "c2"},
	), Commentf("%v", row))

}
