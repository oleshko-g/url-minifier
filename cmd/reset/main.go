package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"io"
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

// generateResetMethods generates "reset.gen.go" file in the path of pkg.
// "reset.gen.go" contains [Reset] method for each struct that's been declared in the package scope with the above "// generate:reset" comment.
func generateResetMethods(pkg *packages.Package) error {
	structTypeDecls := findStructTypeDeclsToReset(pkg.Syntax)
	if structTypeDecls == nil {
		return nil
	}

	structTypes := mapToMap(
		func(structTypeDecl *ast.StructType) *types.Struct {
			return pkg.TypesInfo.Types[structTypeDecl].Type.(*types.Struct)
		}, structTypeDecls)

	if structTypes != nil {
		b := bytes.Buffer{}
		pkgName := pkg.Name
		fmt.Printf("generating \"reset.get.go\" for pakage: %s", pkgName)
		templ, err := template.New("pkg").Parse(pkgTmpl)
		if err != nil {
			return err
		}

		err = templ.Execute(&b, pkg.Name)
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

func generateResetMethod(wr io.Writer, name string, fields []*ast.Field) error {
	templ, err := template.New("resetMeth").Parse(resetMethTmpl)
	if err != nil {
		return err
	}

	resetMeth := struct {
		RecName    string
		StructName string
	}{
		RecName:    strings.ToLower(name[:1]),
		StructName: name,
	}

	err = templ.Execute(wr, resetMeth)
	if err != nil {
		return err
	}

	_, _ = wr, fields
	return nil
}

// structTypesToReset returns struct types marked with `// generate:reset` comment.

// resetStructs return a [ResetStruct] slice to feed into a template to generate the Reset method
func resetStructs(structTypes map[string]*ast.StructType) []ResetStruct {
	var resetStructs []ResetStruct

	for name, structType := range structTypes {

		resetStructs = append(resetStructs, ResetStruct{
			Name:             name,
			FieldsByResetWay: fieldsByResetWay(structType.Fields),
		})
	}

	return resetStructs
}

func fieldsByResetWay(fields *ast.FieldList) map[way][]string {
	fieldsByResetWay := make(map[way][]string)

	for _, field := range fields.List {
		// * TODO: lookup in types info and determine the basic type and

		if fieldType, ok := field.Type.(types.Type); ok {
			underlyingType := underlyingTypeOf(fieldType)

			if basicType := isBasicType(underlyingType); basicType != nil {
				if isScalar(basicType.Kind()) {
					fmt.Println(field.Type.(*ast.Ident).Name)
				}
			}
			_ = underlyingType

		}

		fieldTypeName := field.Type.(*ast.Ident).Name
		// * TODO: determine the reset way
		resetWay := resetWay(fieldTypeName)

		var fieldNames []string
		for _, fieldIdent := range field.Names {
			fieldNames = append(fieldNames, fieldIdent.Name)
		}

		fieldsByResetWay[resetWay] = append(fieldsByResetWay[resetWay], fieldNames...)
	}
	return fieldsByResetWay
}

func underlyingTypeOf(t types.Type) types.Type {
	var underlying types.Type
	for {
		underlying = t.Underlying()
		if underlying == t {
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

func resetWay(fieldTypeName string) way {

	switch fieldTypeName {
	default:
		return ""
	}

}

func isBasicType(t types.Type) *types.Basic {
	if v, ok := t.(*types.Basic); ok {
		return v
	}
	return nil
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

type ResetStruct struct {
	Name             string
	FieldsByResetWay map[way][]string
}

type way string

func group[E any, K comparable, V any](slice []E, f func(E) (k K, vv []V)) map[K][]V {

	var m = make(map[K][]V)

	for _, v := range slice {
		if k, vv := f(v); vv != nil {
			m[k] = append(m[k], vv...)
		}
	}

	return m
}

var (
	pkgTmpl = `// Code generated by go generate; DO NOT EDIT.
// generated by /cmd/reset

package {{.}}

type Resetter interface {
	Reset()
}

type scalar interface {
	~bool | ~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~complex64 | ~complex128
}

type zeroer[T scalar] struct{}

func (z zeroer[T]) zero() T {
	var v T
	return v
}

func truncate[T any](v []T) []T {
	return v[:0]
}

`

	resetMethTmpl = `func ({{.RecName}} *{{.StructName}}) Reset() {
	if {{.RecName}} == nil {
		return
	}

}`
)

func sliceToSlice[T1 any, T2 any](sliceOf []T1, transform func(from T1) (to T2)) []T2 {
	var result = make([]T2, len(sliceOf))
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
