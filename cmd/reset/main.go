package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/tools/go/packages"
)

func main() {
	dirPath := os.Args[1]

	pkgs, err := loadPackages(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range pkgs {
		fmt.Println(pkg.Name)
		for i, _ := range pkg.Syntax {
			fmt.Println(pkg.CompiledGoFiles[i])
		}
	}
}

func loadPackages(dirPath string) ([]*packages.Package, error) {
	pkgs, err := packages.Load(&packages.Config{
		Dir:  dirPath,
		Mode: packages.NeedName | packages.LoadSyntax | packages.LoadFiles,
	}, "./...")
	if err != nil {
		return nil, err
	}

	return pkgs, nil
}
