package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func main() {
	dir := os.Getenv("GOPATH") + "/src/el/services/"
	if err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() &&
				strings.Contains(info.Name(), ".go") &&
				!strings.Contains(info.Name(), "_test") {
				if api, err := newAPI(path); err == nil {
					fmt.Println(api.toLeagueWebSocketPayload())
				}
			}
			return nil
		}); err != nil {
		log.Println(err)
	}
}

//TODO: should be a tree in the future
type API struct {
	Name   string
	Params string //json format
}

func (api *API) toLeagueWebSocketPayload() string {
	return fmt.Sprintf("{\"message_type\":%s,\"info\":{%s}}", api.Name, api.Params)
}

func newAPI(fileName string) (*API, error) {
	f, err := parser.ParseFile(token.NewFileSet(), fileName, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var name string
	var params []string

	//build API based on league conventions
	ast.Inspect(f, func(node ast.Node) bool {
		switch x := node.(type) {
		case *ast.CallExpr:
			selector, ok := x.Fun.(*ast.SelectorExpr)
			if ok && strings.Contains(selector.Sel.Name, "Register") {
				basic, ok := x.Args[0].(*ast.BasicLit)
				if ok {
					name = basic.Value
					return false
				}
			}
		case *ast.TypeSpec:
			if x.Name.Name == "Request" {
				s, ok := x.Type.(*ast.StructType)
				if ok {
					for _, field := range s.Fields.List {
						if field.Tag != nil {
							tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
							paramName := tag.Get("json")
							paramSample := tag.Get("sample")
							params = append(params, fmt.Sprintf("\"%s\":\"%s\"", paramName, paramSample))
						}
					}
					return false
				}
			}
		}
		return true
	})

	if name == "" {
		return nil, fmt.Errorf("api name cannot be empty")
	}

	return &API{
		Name:   name,
		Params: strings.Join(params, ","),
	}, nil
}
