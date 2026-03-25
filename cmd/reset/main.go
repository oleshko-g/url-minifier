package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"iter"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

func main() {
	// dirPath := os.Args[:1]
	dirPath := "../../"

	pkgs, err := loadPackages(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range pkgs {
		err = generateResetMethods(pkg)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func loadPackages(dirPath string) ([]*packages.Package, error) {
	pkgs, err := packages.Load(&packages.Config{
		Dir:  dirPath,
		Mode: packages.LoadSyntax,
	}, "./...")
	if err != nil {
		return nil, err
	}

	return pkgs, nil
}

type ResetData struct {
	PkgName      string
	ResetStructs []ResetStruct
}

// generateResetMethods generates "reset.gen.go" file in the path of pkg.
// "reset.gen.go" contains [Reset] method for each struct that's been declared in the package scope with the above "// generate:reset" comment.
func generateResetMethods(pkg *packages.Package) error {
	structTypeDecls := findStructTypeDeclsToReset(pkg.Syntax)
	if structTypeDecls == nil {
		return nil
	}

	structTypes := mapToMap(
		func(structTypeDecl *ast.StructType) *types.Struct {
			// find [types.Type] by declaration and assert to [types.Struct]
			return pkg.TypesInfo.Types[structTypeDecl].Type.(*types.Struct)
		}, structTypeDecls)

	resetStructs := resetStructs(structTypes)
	if resetStructs != nil {
		fmt.Printf("generating \"reset.get.go\" for pakage: %s", pkg.Name)
		templ, err := template.New("pkg").Parse(pkgTmpl)
		if err != nil {
			return err
		}

		b := bytes.Buffer{}
		resetData := ResetData{
			PkgName:      pkg.Name,
			ResetStructs: resetStructs,
		}
		err = templ.Execute(&b, resetData)
		if err != nil {
			return err
		}

		path := path.Join(pkg.Dir, "reset.gen.go")

		err = os.WriteFile(path, b.Bytes(), 0o755)
		if err != nil {
			return err
		}
	}

	return nil
}

// findStructTypeDeclsToReset returns struct type declarations marked with the `// generate:reset` comment.
func findStructTypeDeclsToReset(pkgFiles []*ast.File) (structTypes map[*ast.Ident]*ast.StructType) {
	for _, file := range pkgFiles {
		var genDecls []*ast.GenDecl
		for _, v := range file.Decls {
			if genDecl, ok := v.(*ast.GenDecl); ok {
				genDecls = append(genDecls, genDecl)
			}
		}

		for _, genDecl := range genDecls {
			if hasComment("// generate:reset", genDecl) {
				if ident, structType := structFromGenDecl(genDecl); structType != nil {
					// if a first find allocate a map
					if structTypes == nil {
						structTypes = make(map[*ast.Ident]*ast.StructType)
					}

					structTypes[ident] = structType
				}
			}
		}
	}
	return structTypes
}

type ResetStruct struct {
	Rcv, Name        string
	FieldsByResetWay map[way][]ResetField
}

// resetStructs return a [ResetStruct] slice to feed into a template to generate the Reset method
func resetStructs(structTypes map[*ast.Ident]*types.Struct) []ResetStruct {
	resetStructs := mapToSlice(structTypes,
		func(ident *ast.Ident, structType *types.Struct) ResetStruct {
			return ResetStruct{
				Rcv:              strings.ToLower(ident.Name[:1]),
				Name:             ident.Name,
				FieldsByResetWay: fieldsByResetWay(structType.Fields()),
			}
		})

	return resetStructs
}

type ResetField struct {
	TypeName, FieldName string
}

func fieldsByResetWay(fields iter.Seq[*types.Var]) map[way][]ResetField {
	fieldsByResetWay := make(map[way][]ResetField)

	for field := range fields {
		underlyingType := underlyingTypeOf(field.Type())

		resetWay := resetWay(underlyingType)
		if resetWay == "unsupported" {
			continue
		}

		fieldsByResetWay[resetWay] = append(fieldsByResetWay[resetWay], ResetField{
			// type name relative to the field package
			TypeName:  types.TypeString(field.Type(), types.RelativeTo(field.Pkg())),
			FieldName: field.Name(),
		})

	}

	return fieldsByResetWay
}

func underlyingTypeOf(t types.Type) types.Type {
	underlying := t.Underlying()

	for {
		if underlying == underlying.Underlying() {
			return underlying
		}
	}
}

func isScalar(t types.BasicKind) bool {
	if t != types.UnsafePointer &&
		t != types.UntypedNil &&
		t != types.Rune &&
		t != types.Byte {
		return true
	}

	return false
}

type way string

func resetWay(t types.Type) way {
	switch v := t.(type) {
	case *types.Pointer:
		if resetWay(v.Elem()) == "scalar" {
			return way("scalarPtr")
		}
		return way("ptr")
	case *types.Map:
		return way("map")
	case *types.Slice:
		return way("slice")
	case *types.Basic:
		if isScalar(v.Kind()) {
			return way("scalar")
		}
	case *types.Struct:
		return way("struct")
	}

	return "unsupported"
}

// structFromGenDecl transforms [*ast.GenDecl] to [*ast.StructType] declaration.
func structFromGenDecl(decl *ast.GenDecl) (name *ast.Ident, typ *ast.StructType) {
	if len(decl.Specs) == 1 {
		if typeSpec, ok := decl.Specs[0].(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				return typeSpec.Name, structType
			}
		}
	}

	return nil, nil
}

func hasComment(comment string, decl *ast.GenDecl) bool {
	if decl.Doc != nil {
		for _, doc := range decl.Doc.List {
			if doc.Text == comment {
				return true
			}
		}
	}

	return false
}

var pkgTmpl = `// Code generated by go generate; DO NOT EDIT.
// generated by /cmd/reset

package {{.PkgName}}

type Resetter interface {
	Reset()
}

type scalar interface {
	~bool | ~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~complex64 | ~complex128
}

func zero[S scalar]() S {
	var v S
	return v
}

func truncate[T any](v []T) []T {
	return v[:0]
}

{{range $resetStruct := .ResetStructs -}}

// Reset sets the {{.Name}}
func ({{.Rcv}} *{{.Name}}) Reset() {
	if {{.Rcv}} == nil {
		return
	}
{{range $resetWay, $fields := .FieldsByResetWay -}}

{{if eq $resetWay "map"}}
	{{range $fields -}}
		clear({{$resetStruct.Rcv}}.{{.FieldName}})
	{{end -}}
{{end -}}

{{if eq $resetWay "scalar"}}
	{{range $fields -}}
		{{- $resetStruct.Rcv}}.{{.FieldName}} = zero[{{.TypeName}}]()
	{{end -}}
{{end -}}

{{- if eq $resetWay "scalarPtr"}}
	{{- range $field := $fields}}
	{{with $baseTypeName := slice $field.TypeName 1 -}}
	*{{- $resetStruct.Rcv}}.{{$field.FieldName}} = zero[{{$baseTypeName}}]()
	{{end -}}
	{{end -}}
{{end -}}

{{- if eq $resetWay "slice"}}
	{{range $fields}}
		{{- $resetStruct.Rcv}}.{{.FieldName}} = truncate({{$resetStruct.Rcv}}.{{.FieldName}})
	{{end -}}
{{end -}}

{{- if eq $resetWay "struct"}}
	{{- range $fields}}
	if resetter, ok := any({{$resetStruct.Rcv}}.{{.FieldName}}).(Resetter); ok {
	    resetter.Reset()
	}
	{{end -}}
{{end -}}

{{- end}}
}{{end}}`

func sliceToSlice[T1 any, T2 any](sliceOf []T1, transform func(from T1) (to T2)) []T2 {
	result := make([]T2, len(sliceOf))
	for i, v := range sliceOf {
		result[i] = transform(v)
	}

	return result
}

// mapToMap transform a KV1 map to a KV2 map
func mapToMap[K comparable, V1 any, V2 any](transform func(V1) V2, from map[K]V1) (to map[K]V2) {
	to = make(map[K]V2)

	for k, v := range from {
		to[k] = transform(v)
	}
	return to
}

func toFilter[T any](sliceOf []T, meets func(T) bool) (subSliceOf []T) {
	for _, v := range sliceOf {
		if meets(v) {
			subSliceOf = append(subSliceOf, v)
		}
	}

	return subSliceOf
}

func mapToSlice[M map[K]V1, K comparable, V1 any, V2 any](fromMap M, transform func(K, V1) V2) (sliceOf []V2) {
	sliceOf = make([]V2, len(fromMap))

	var i int
	for k, v := range fromMap {
		sliceOf[i] = transform(k, v)
		i++
	}

	return sliceOf
}

func group[E any, K comparable, V any](slice []E, f func(E) (k K, vv []V)) map[K][]V {
	m := make(map[K][]V)

	for _, v := range slice {
		if k, vv := f(v); vv != nil {
			m[k] = append(m[k], vv...)
		}
	}

	return m
}
