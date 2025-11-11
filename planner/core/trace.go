package core

import (
	"github.com/sqllabs/mysqlaudit/ast"
)

// Trace represents a trace plan.
type Trace struct {
	baseSchemaProducer

	StmtNode ast.StmtNode
}
