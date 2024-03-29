//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var ParentsChildren = newParentsChildrenTable("public", "parents_children", "")

type parentsChildrenTable struct {
	postgres.Table

	//Columns
	ParentID postgres.ColumnInteger
	ChildID  postgres.ColumnInteger

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type ParentsChildrenTable struct {
	parentsChildrenTable

	EXCLUDED parentsChildrenTable
}

// AS creates new ParentsChildrenTable with assigned alias
func (a ParentsChildrenTable) AS(alias string) *ParentsChildrenTable {
	return newParentsChildrenTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new ParentsChildrenTable with assigned schema name
func (a ParentsChildrenTable) FromSchema(schemaName string) *ParentsChildrenTable {
	return newParentsChildrenTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new ParentsChildrenTable with assigned table prefix
func (a ParentsChildrenTable) WithPrefix(prefix string) *ParentsChildrenTable {
	return newParentsChildrenTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new ParentsChildrenTable with assigned table suffix
func (a ParentsChildrenTable) WithSuffix(suffix string) *ParentsChildrenTable {
	return newParentsChildrenTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newParentsChildrenTable(schemaName, tableName, alias string) *ParentsChildrenTable {
	return &ParentsChildrenTable{
		parentsChildrenTable: newParentsChildrenTableImpl(schemaName, tableName, alias),
		EXCLUDED:             newParentsChildrenTableImpl("", "excluded", ""),
	}
}

func newParentsChildrenTableImpl(schemaName, tableName, alias string) parentsChildrenTable {
	var (
		ParentIDColumn = postgres.IntegerColumn("parent_id")
		ChildIDColumn  = postgres.IntegerColumn("child_id")
		allColumns     = postgres.ColumnList{ParentIDColumn, ChildIDColumn}
		mutableColumns = postgres.ColumnList{}
	)

	return parentsChildrenTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ParentID: ParentIDColumn,
		ChildID:  ChildIDColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
