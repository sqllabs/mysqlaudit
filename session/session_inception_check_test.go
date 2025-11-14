// Copyright (C) 2025 JustCoding247. All rights reserved.
package session

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/sqllabs/mysqlaudit/ast"
	"github.com/sqllabs/mysqlaudit/parser"
)

func newTestSession() *session {
	return &session{
		stage:          StageCheck,
		recordSets:     NewRecordSets(),
		myRecord:       &Record{Stage: StageCheck, Buf: new(bytes.Buffer)},
		opt:            &SourceOptions{Execute: true},
		tableCacheList: make(map[string]*TableInfo),
		dbName:         "test",
	}
}

func cachedTable(s *session, schema, name string) *TableInfo {
	key := fmt.Sprintf("%s.%s", schema, name)
	if tbl, ok := s.tableCacheList[key]; ok {
		return tbl
	}
	return nil
}

func TestAttachColumnCheckConstraint(t *testing.T) {
	s := newTestSession()
	p := parser.New()
	stmt, err := p.ParseOneStmt("CREATE TABLE t (age INT CHECK (age >= 18))", "", "")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	createStmt := stmt.(*ast.CreateTableStmt)
	col := createStmt.Cols[0]

	table := &TableInfo{Name: "t"}
	s.attachColumnCheckConstraints(table, col.Name.Name.O, col.Options)

	if len(table.Checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(table.Checks))
	}
	chk := table.Checks[0]
	if chk.Level != "COLUMN" || chk.ColumnName != "age" {
		t.Fatalf("unexpected column-level metadata: %+v", chk)
	}
	if chk.Expression != "`age`>=18" {
		t.Fatalf("unexpected expression: %s", chk.Expression)
	}
	if !chk.Enforced {
		t.Fatalf("column checks should default to enforced")
	}
	msg := s.myRecord.Buf.String()
	if !strings.Contains(msg, "COLUMN `age` CHECK constraint (unnamed)") {
		t.Fatalf("missing info message, got %s", msg)
	}
}

func TestAddTableCheckConstraintNotEnforced(t *testing.T) {
	s := newTestSession()
	p := parser.New()
	sql := "CREATE TABLE t (id INT, CONSTRAINT chk_positive CHECK (id > 0) NOT ENFORCED)"
	stmt, err := p.ParseOneStmt(sql, "", "")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	createStmt := stmt.(*ast.CreateTableStmt)
	if len(createStmt.Constraints) != 1 {
		t.Fatalf("expected 1 constraint, got %d", len(createStmt.Constraints))
	}

	table := &TableInfo{Name: "t"}
	s.addTableCheckConstraint(table, createStmt.Constraints[0], "TABLE `t`")

	if len(table.Checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(table.Checks))
	}
	chk := table.Checks[0]
	if chk.Name != "chk_positive" || chk.Level != "TABLE" {
		t.Fatalf("unexpected table-level metadata: %+v", chk)
	}
	if chk.Enforced {
		t.Fatalf("constraint should be NOT ENFORCED")
	}
	if chk.Expression != "`id`>0" {
		t.Fatalf("expression mismatch: %s", chk.Expression)
	}
	if s.myRecord.ErrLevel != 1 {
		t.Fatalf("warning level not raised, errLevel=%d", s.myRecord.ErrLevel)
	}
	msg := s.myRecord.Buf.String()
	if !strings.Contains(msg, "[NOT ENFORCED]") {
		t.Fatalf("expected warning message, got %s", msg)
	}
}

func TestCheckConstraintRollbackHelpers(t *testing.T) {
	s := newTestSession()
	info := CheckConstraintInfo{
		Name:       "chk_age",
		Expression: "`age`>=18",
		Enforced:   true,
	}

	s.mysqlAddCheckRollback("chk_age")
	s.mysqlDropCheckRollback(info)
	s.mysqlAlterCheckRollback("chk_age", false)

	if len(s.alterRollbackBuffer) != 3 {
		t.Fatalf("unexpected rollback buffer size %d", len(s.alterRollbackBuffer))
	}
	if s.alterRollbackBuffer[0] != "DROP CHECK `chk_age`," {
		t.Fatalf("unexpected add rollback: %s", s.alterRollbackBuffer[0])
	}
	if s.alterRollbackBuffer[1] != "ADD CONSTRAINT `chk_age` CHECK (`age`>=18)," {
		t.Fatalf("unexpected drop rollback: %s", s.alterRollbackBuffer[1])
	}
	if s.alterRollbackBuffer[2] != "ALTER CHECK `chk_age` NOT ENFORCED," {
		t.Fatalf("unexpected alter rollback: %s", s.alterRollbackBuffer[2])
	}
}

func TestColumnCheckLifecycleHelpers(t *testing.T) {
	table := &TableInfo{
		Checks: []CheckConstraintInfo{
			{Name: "c_age", ColumnName: "age"},
			{Name: "c_salary", ColumnName: "salary"},
		},
	}
	removed := table.RemoveColumnChecks("age")
	if len(removed) != 1 || removed[0].Name != "c_age" {
		t.Fatalf("remove column checks failed: %+v", removed)
	}
	if len(table.Checks) != 1 || table.Checks[0].Name != "c_salary" {
		t.Fatalf("remaining checks incorrect: %+v", table.Checks)
	}
	table.RenameColumnChecks("salary", "pay")
	if table.Checks[0].ColumnName != "pay" {
		t.Fatalf("rename column checks failed: %+v", table.Checks[0])
	}
}

func TestAlterTableAlterCheck(t *testing.T) {
	s := newTestSession()
	table := &TableInfo{
		Schema: "test",
		Name:   "t",
		Checks: []CheckConstraintInfo{
			{Name: "chk_age", Expression: "`age`>0", Enforced: true},
		},
	}
	spec := &ast.AlterTableSpec{
		Tp: ast.AlterTableAlterCheck,
		Constraint: &ast.Constraint{
			Name:     "chk_age",
			Enforced: false,
		},
	}

	s.checkAlterTableAlterCheck(table, spec)

	result := cachedTable(s, "test", "t")
	if result == nil {
		t.Fatalf("table snapshot not cached")
	}
	if len(result.Checks) != 1 || result.Checks[0].Name != "chk_age" {
		t.Fatalf("constraint missing after ALTER CHECK: %+v", result.Checks)
	}
	if result.Checks[0].Enforced {
		t.Fatalf("constraint should be marked NOT ENFORCED")
	}
	if len(s.alterRollbackBuffer) != 1 || s.alterRollbackBuffer[0] != "ALTER CHECK `chk_age` ENFORCED," {
		t.Fatalf("unexpected rollback buffer: %#v", s.alterRollbackBuffer)
	}
}

func TestAlterTableDropCheck(t *testing.T) {
	s := newTestSession()
	table := &TableInfo{
		Schema: "test",
		Name:   "t",
		Checks: []CheckConstraintInfo{
			{Name: "chk_age", Expression: "`age`>0", Enforced: true},
		},
	}
	spec := &ast.AlterTableSpec{
		Tp: ast.AlterTableDropCheck,
		Constraint: &ast.Constraint{
			Name: "chk_age",
		},
	}

	s.checkAlterTableDropCheck(table, spec)

	result := cachedTable(s, "test", "t")
	if result == nil {
		t.Fatalf("table snapshot not cached")
	}
	if len(result.Checks) != 0 {
		t.Fatalf("constraint not removed: %+v", result.Checks)
	}
	if len(s.alterRollbackBuffer) != 1 {
		t.Fatalf("unexpected rollback buffer: %#v", s.alterRollbackBuffer)
	}
	if !strings.Contains(s.alterRollbackBuffer[0], "ADD CONSTRAINT `chk_age` CHECK (`age`>0)") {
		t.Fatalf("drop rollback SQL incorrect: %s", s.alterRollbackBuffer[0])
	}
	msg := s.myRecord.Buf.String()
	if !strings.Contains(msg, "DROP CHECK `chk_age`") {
		t.Fatalf("expected info message about DROP CHECK, got %s", msg)
	}
}

func TestUpdateColumnCheckConstraints(t *testing.T) {
	s := newTestSession()
	table := &TableInfo{
		Name: "users",
		Checks: []CheckConstraintInfo{
			{ColumnName: "age", Expression: "`age`>0", Level: "COLUMN"},
		},
	}
	opts := parseColumnOptions(t, "CREATE TABLE users (age INT CHECK (age >= 18))", "age")

	s.updateColumnCheckConstraints(table, "age", opts)

	if len(table.Checks) != 1 {
		t.Fatalf("expected 1 check after update, got %d", len(table.Checks))
	}
	if table.Checks[0].Expression != "`age`>=18" {
		t.Fatalf("check expression not updated: %s", table.Checks[0].Expression)
	}
	if table.Checks[0].ColumnName != "age" {
		t.Fatalf("column name changed unexpectedly: %s", table.Checks[0].ColumnName)
	}
}

func TestChangeColumnRenameChecks(t *testing.T) {
	s := newTestSession()
	table := &TableInfo{
		Name: "users",
		Checks: []CheckConstraintInfo{
			{ColumnName: "age", Expression: "`age`>0", Level: "COLUMN"},
		},
	}

	table.RenameColumnChecks("age", "user_age")
	opts := parseColumnOptions(t, "CREATE TABLE users (user_age INT CHECK (user_age <= 100))", "user_age")

	s.updateColumnCheckConstraints(table, "user_age", opts)

	if len(table.Checks) != 1 {
		t.Fatalf("expected 1 check after rename/update, got %d", len(table.Checks))
	}
	chk := table.Checks[0]
	if chk.ColumnName != "user_age" {
		t.Fatalf("column name not updated: %s", chk.ColumnName)
	}
	if chk.Expression != "`user_age`<=100" {
		t.Fatalf("check expression not updated: %s", chk.Expression)
	}
}

func parseColumnOptions(t *testing.T, sql, column string) []*ast.ColumnOption {
	t.Helper()
	p := parser.New()
	stmt, err := p.ParseOneStmt(sql, "", "")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	create, ok := stmt.(*ast.CreateTableStmt)
	if !ok {
		t.Fatalf("expected CreateTableStmt, got %T", stmt)
	}
	for _, col := range create.Cols {
		if strings.EqualFold(col.Name.Name.O, column) {
			return col.Options
		}
	}
	t.Fatalf("column %s not found in SQL", column)
	return nil
}
