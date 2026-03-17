package main

import (
	"fmt"
	"go/ast"
	"log"

	"os"

	"golang.org/x/tools/go/packages"
)

func main() {
	dirPath := os.Args[1]
	// dirPath := "../../"

	pkgs, err := loadPackages(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range pkgs {
		fmt.Println(pkg.Name)
		err = generateResetMethods(pkg)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func loadPackages(dirPath string) ([]*packages.Package, error) {
	pkgs, err := packages.Load(&packages.Config{
		Dir:  dirPath,
		Mode: packages.NeedName | packages.LoadSyntax | packages.NeedModule,
	}, "./...")
	if err != nil {
		return nil, err
	}

	return pkgs, nil
}

func generateResetMethods(pkg *packages.Package) error {
	for _, file := range pkg.Syntax {
		structs := structsToReset(file)
		if structs == nil {
			return nil
		}
	}

	return nil
}

func structsToReset(file *ast.File) map[string]*ast.StructType {
	var structsToReset = make(map[string]*ast.StructType)

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if ok {
			if hasComment("// generate:reset", genDecl) {
				if name, structType := isStructDecl(genDecl); structType != nil {
					structsToReset[name] = structType
				}
			}
		}
	}

	return structsToReset
}

func isStructDecl(decl *ast.GenDecl) (name string, typ *ast.StructType) {
	if len(decl.Specs) == 1 {
		if typeSpec, ok := decl.Specs[0].(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				return typeSpec.Name.Name, structType
			}
		}
	}

	return "", nil
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
