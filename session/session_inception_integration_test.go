// Copyright (C) 2025 JustCoding247. All rights reserved.
//go:build integration
// +build integration

package session

import (
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

func TestQueryCheckConstraintsFromDBIntegration(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	schema := os.Getenv("MYSQL_TEST_DB")
	if dsn == "" || schema == "" {
		t.Skip("MYSQL_TEST_DSN/MYSQL_TEST_DB not set")
	}

	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to connect test database: %v", err)
	}
	defer db.Close()

	tableName := "check_constraint_integration"
	dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS `%s`.`%s`", schema, tableName)
	if err := db.Exec(dropSQL).Error; err != nil {
		t.Fatalf("failed to drop test table: %v", err)
	}
	createSQL := fmt.Sprintf("CREATE TABLE `%s`.`%s` (id INT, CONSTRAINT `chk_positive` CHECK (id > 0) NOT ENFORCED)", schema, tableName)
	if err := db.Exec(createSQL).Error; err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}
	defer db.Exec(dropSQL)

	s := newTestSession()
	s.db = db
	s.dbName = schema

	checks := s.queryCheckConstraintsFromDB(schema, tableName)
	if len(checks) == 0 {
		t.Fatalf("expected at least one check constraint, got %d", len(checks))
	}
	found := false
	for _, chk := range checks {
		if chk.Name == "chk_positive" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("constraint chk_positive not found: %+v", checks)
	}
}
