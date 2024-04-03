package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/printer"
	"go/token"
	"io"
	"os"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

// turn openapi.Collection into a terraform resource model using the ast

//go:generate go run ../generator/cmd/main.go openapi.Collection

type ResourceModel struct {
	Package        string
	Name           string
	Model          string
	FieldOverrides map[string]string
}

func New(m string) *ResourceModel {
	return &ResourceModel{
		Model:          m,
		FieldOverrides: map[string]string{},
	}
}

func (g ResourceModel) Load(name string) error {
	fmt.Printf("loading package %s\n", name)
	// model := reflect.New(reflect.Y)
	ctx := build.Default
	// ctx.Dir = "."
	pkg, err := ctx.Import(name, ".", build.FindOnly)
	if err != nil {
		return fmt.Errorf("failed to import package %s: %w", name, err)
	}

	fmt.Printf("imported package %s: %s\n", pkg.Name, pkg.ImportPath)

	p, err := packages.Load(&packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:   ".",
		Tests: true,
	}, pkg.ImportPath)
	if err != nil {
		return fmt.Errorf("failed to load package %s: %w", name, err)
	}
	fmt.Printf("loaded %d packages\n", len(p))

	return nil
}

func (g ResourceModel) Generate(out io.WriteCloser) error {
	file := ast.File{
		Name: ast.NewIdent(g.Package),
	}
	genDecl := ast.GenDecl{Tok: token.TYPE}
	typeSpec := ast.TypeSpec{Name: ast.NewIdent(g.Name + "ResourceModel")}
	structType := ast.StructType{Fields: &ast.FieldList{}}

	fields := reflect.VisibleFields(reflect.TypeOf(g.Model))
	for _, field := range fields {
		tag := field.Tag.Get("json")
		name := field.Name
		if n, ok := g.FieldOverrides[name]; ok {
			name = n
		}
		fieldType := field.Type

		structType.Fields.List = append(structType.Fields.List,
			&ast.Field{
				Names: []*ast.Ident{ast.NewIdent(name)},
				Type:  ast.NewIdent(fieldType.String()),
				Tag:   tfTag(tag),
			})
	}

	typeSpec.Type = &structType
	genDecl.Specs = append(genDecl.Specs, &typeSpec)
	file.Decls = append(file.Decls, &genDecl)

	buf := make([]byte, 0, 4096)
	b := bytes.NewBuffer(buf)
	err := printer.Fprint(b, token.NewFileSet(), &file)
	if err != nil {
		return fmt.Errorf("failed to generate file: %w", err)
	}
	if _, err = fmt.Fprintf(os.Stdout, "%d: %s\n", len(buf), string(buf)); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	b2, err := format.Source(buf)
	if err != nil {
		return fmt.Errorf("failed to format generated file: %w", err)
	}

	if _, err = fmt.Fprintf(os.Stdout, "%d: %s\n", len(b2), string(b2)); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	return nil
}

func tfTag(tag string) *ast.BasicLit {
	fields := strings.Split(tag, ",")
	for _, f := range fields {
		switch f {
		case "omitempty", "-":
		default:
			return &ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf(`tfsdk:"%s"`, f),
			}
		}
	}

	panic("no tag found")
}
