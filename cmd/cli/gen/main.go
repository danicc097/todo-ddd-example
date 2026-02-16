package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stoewer/go-strcase"
)

type PathParam struct {
	Name   string
	GoType string
}

type ApiParam struct {
	GoName      string
	FlagName    string
	Description string
	IsString    bool
	IsInt       bool
	IsUUID      bool
	Required    bool
}

type CommandData struct {
	OperationID       string
	PascalOperationID string
	Use               string
	Short             string
	HasBody           bool
	BodyType          string
	PathParams        []PathParam
	ApiParams         []ApiParam
	HasParams         bool
}

func main() {
	loader := openapi3.NewLoader()

	doc, err := loader.LoadFromFile("../../openapi.yaml")
	if err != nil {
		log.Fatalf("Failed to load spec: %v", err)
	}

	var commands []CommandData

	paths := make([]string, 0, len(doc.Paths.Map()))
	for p := range doc.Paths.Map() {
		paths = append(paths, p)
	}

	sort.Strings(paths) // ensure idempotent gen

	for _, path := range paths {
		pathItem := doc.Paths.Find(path)
		ops := pathItem.Operations()

		methods := make([]string, 0, len(ops))
		for m := range ops {
			methods = append(methods, m)
		}

		sort.Strings(methods) // ensure idempotent gen

		for _, method := range methods {
			op := ops[method]
			if op.OperationID == "" {
				continue
			}

			pascalOpID := op.OperationID
			if len(pascalOpID) > 0 {
				pascalOpID = strings.ToUpper(pascalOpID[:1]) + pascalOpID[1:]
			}

			cmd := CommandData{
				OperationID:       op.OperationID,
				PascalOperationID: pascalOpID,
				Use:               strcase.KebabCase(op.OperationID),
				Short:             op.Summary,
			}

			for _, paramRef := range op.Parameters {
				param := paramRef.Value
				if param.In == "path" {
					goType := "uuid.UUID"

					if ext, ok := param.Schema.Value.Extensions["x-go-type"]; ok {
						switch v := ext.(type) {
						case string:
							goType = strings.Trim(v, "\"")
						case json.RawMessage:
							goType = strings.Trim(string(v), "\"")
						}
					}

					cmd.PathParams = append(cmd.PathParams, PathParam{
						Name:   param.Name,
						GoType: goType,
					})
					cmd.Use += fmt.Sprintf(" [%s]", param.Name)
				} else {
					cmd.HasParams = true

					// adhoc initialism fixes to match oapi-codegen
					goName := strcase.UpperCamelCase(param.Name)

					goName = strings.ReplaceAll(goName, "Otp", "OTP")
					if !strings.HasPrefix(goName, "Idemp") {
						goName = strings.ReplaceAll(goName, "Id", "ID")
					}

					apiParam := ApiParam{
						GoName:      goName,
						FlagName:    strcase.KebabCase(param.Name),
						Description: strings.ReplaceAll(param.Description, "\n", " "),
						Required:    param.Required,
					}

					schemaTypeString := false
					schemaTypeInt := false

					if param.Schema != nil && param.Schema.Value != nil {
						if param.Schema.Value.Type != nil {
							schemaTypeString = param.Schema.Value.Type.Includes("string")
							schemaTypeInt = param.Schema.Value.Type.Includes("integer")
						}

						if param.Schema.Value.Format == "uuid" {
							apiParam.IsUUID = true
						}
					}

					apiParam.IsString = schemaTypeString && !apiParam.IsUUID
					apiParam.IsInt = schemaTypeInt
					cmd.ApiParams = append(cmd.ApiParams, apiParam)
				}
			}

			if op.RequestBody != nil {
				cmd.HasBody = true
				cmd.BodyType = pascalOpID + "JSONRequestBody"
			}

			commands = append(commands, cmd)
		}
	}

	tmpl, err := template.ParseFiles("gen/cli.gotmpl")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, commands); err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalf("Failed to format code: %v\n%s", err, buf.String())
	}

	if err := os.WriteFile("commands.gen.go", formatted, 0o644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
}
