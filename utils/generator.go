package utils

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type GeneratorConfig struct {
	Dir           string
	OutFile       string
	VarName       string
	SkipEmbedded  bool
	LowercaseKeys bool
}

func generateGettersToWriter(w io.Writer, pkgs map[string]*ast.Package, pkgName string, config GeneratorConfig) error {
	fmt.Fprintf(w, "package %s\n\n", pkgName)
	fmt.Fprintf(w, "var %s = map[string]any{\n", config.VarName)

	// First pass: collect all struct types
	structTypes := make(map[string]*ast.StructType)
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			for _, decl := range f.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				for _, spec := range gen.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					st, ok := ts.Type.(*ast.StructType)
					if !ok {
						continue
					}
					structTypes[ts.Name.Name] = st
				}
			}
		}
	}

	// Second pass: generate getters with nested field support
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			for _, decl := range f.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				for _, spec := range gen.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					st, ok := ts.Type.(*ast.StructType)
					if !ok {
						continue
					}
					structName := ts.Name.Name

					fmt.Fprintf(w, "\t%q: map[string]func(*%s)any{\n", structName, structName)
					for _, field := range st.Fields.List {
						if len(field.Names) == 0 {
							if config.SkipEmbedded {
								continue // Skip embedded fields
							}
						} else {
							fieldName := field.Names[0].Name
							key := fieldName

							// Get JSON tag if it exists
							if field.Tag != nil {
								tag := field.Tag.Value
								// Remove backticks
								tag = strings.Trim(tag, "`")
								// Parse json tag
								if strings.Contains(tag, "json:") {
									parts := strings.Split(tag, "json:")
									if len(parts) > 1 {
										jsonTag := strings.Trim(parts[1], `"`)
										jsonTag = strings.Split(jsonTag, ",")[0] // Get first part before comma
										if jsonTag != "" && jsonTag != "-" {
											key = jsonTag
										}
									}
								}
							}

							if config.LowercaseKeys && field.Tag == nil {
								key = strings.ToLower(fieldName)
							}

							fmt.Fprintf(w, "\t\t%q: func(u *%s) any { return u.%s },\n",
								key, structName, fieldName)

							// Check if field is a struct type and generate nested getters
							if ident, ok := field.Type.(*ast.Ident); ok {
								if nestedStruct, exists := structTypes[ident.Name]; exists {
									for _, nestedField := range nestedStruct.Fields.List {
										if len(nestedField.Names) > 0 {
											nestedFieldName := nestedField.Names[0].Name
											nestedKey := key + "." + strings.ToLower(nestedFieldName)

											// Get nested JSON tag
											if nestedField.Tag != nil {
												tag := nestedField.Tag.Value
												tag = strings.Trim(tag, "`")
												if strings.Contains(tag, "json:") {
													parts := strings.Split(tag, "json:")
													if len(parts) > 1 {
														jsonTag := strings.Trim(parts[1], `"`)
														jsonTag = strings.Split(jsonTag, ",")[0]
														if jsonTag != "" && jsonTag != "-" {
															nestedKey = key + "." + jsonTag
														}
													}
												}
											}

											fmt.Fprintf(w, "\t\t%q: func(u *%s) any { return u.%s.%s },\n",
												nestedKey, structName, fieldName, nestedFieldName)
										}
									}
								}
							}
						}
					}
					fmt.Fprintln(w, "\t},")
				}
			}
		}
	}
	fmt.Fprintln(w, "}")
	return nil
}

func GenerateFieldGetters(config GeneratorConfig) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, config.Dir, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("error parsing directory: %w", err)
	}
	outFile, err := os.Create(filepath.Join(config.Dir, config.OutFile))
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer outFile.Close()
	pkgName := ""
	for _, pkg := range pkgs {
		pkgName = pkg.Name
		break
	}
	if pkgName == "" {
		return fmt.Errorf("no package found in directory")
	}

	return generateGettersToWriter(outFile, pkgs, pkgName, config)
}

func RunCLI(path string) {
	out := flag.String("out", "gen_field_getters.go", "output file")
	varName := flag.String("var", "ModelFieldGetters", "variable name for the generated map")
	lowercase := flag.Bool("lowercase", true, "use lowercase keys")
	flag.Parse()

	dir := flag.Arg(0)
	if dir == "" {
		dir = path
	}

	config := GeneratorConfig{
		Dir:           dir,
		OutFile:       *out,
		VarName:       *varName,
		SkipEmbedded:  true,
		LowercaseKeys: *lowercase,
	}

	if err := GenerateFieldGetters(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s successfully in %s\n", *out, dir)
}
