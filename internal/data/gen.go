//go:build ignore

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/types"
	"github.com/go-jet/jet/v2/generator/metadata"
	"github.com/go-jet/jet/v2/generator/postgres"
	"github.com/go-jet/jet/v2/generator/template"
	postgres2 "github.com/go-jet/jet/v2/postgres"
	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("MIGRATE_DSN")
	if dsn == "" {
		panic("no DSN")
	}

	err := postgres.GenerateDSN(dsn, "public", "./internal/data/gen", template.Default(postgres2.Dialect).
		UseSchema(func(schemaMetaData metadata.Schema) template.Schema {
			return template.DefaultSchema(schemaMetaData).
				UseModel(template.DefaultModel().
					UseTable(func(table metadata.Table) template.TableModel {
						return template.DefaultTableModel(table).
							UseField(func(columnMetaData metadata.Column) template.TableModelField {
								defaultTableModelField := template.DefaultTableModelField(columnMetaData)

								if table.Name == "users" && columnMetaData.Name == "password" {
									defaultTableModelField.Tags = append(defaultTableModelField.Tags, `json:"-"`)
									defaultTableModelField.Type = template.NewType(new(types.Password))
								} else if table.Name == "sessions" && columnMetaData.Name == "token" {
									defaultTableModelField.Tags = append(defaultTableModelField.Tags, `json:"token"`)
									defaultTableModelField.Type = template.NewType(new(types.Token))
								} else if table.Name == "users" && columnMetaData.Name == "totp_secret" {
									defaultTableModelField.Tags = append(defaultTableModelField.Tags, `json:"-"`)
									defaultTableModelField.Type = template.NewType(new(types.TOTPSecret))
								} else {
									defaultTableModelField.Tags = append(defaultTableModelField.Tags, fmt.Sprintf(`json:"%s,omitempty"`, columnMetaData.Name))
								}

								if table.Name == "assignments" && columnMetaData.Name == "deadline" ||
									table.Name == "users" && columnMetaData.Name == "birth_date" ||
									table.Name == "lessons" && columnMetaData.Name == "date" {
									defaultTableModelField.Type = template.NewType(new(types.Date))
								}

								switch defaultTableModelField.Type.Name {
								case "int32", "*int32":
									if columnMetaData.Name != "id" {
										defaultTableModelField.Type = template.NewType(new(int))
									} else {
										defaultTableModelField.Type = template.NewType(int(1))
									}
								case "string":
									defaultTableModelField.Type = template.NewType(new(string))
								case "time.Time":
									defaultTableModelField.Type = template.NewType(new(time.Time))
								case "bool":
									defaultTableModelField.Type = template.NewType(new(bool))
								}

								return defaultTableModelField
							})
					}))
		}))

	if err != nil {
		panic(fmt.Sprintf("failed to generate: %s", err.Error()))
	}
}
