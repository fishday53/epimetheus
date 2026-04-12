package staticlint

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "reports os.Exit in main()",
	Run:  run,
}

const (
	findPackageArea   = "main"
	findPackageAreaFn = "main"
	findUsedPackage   = "os"
	findUsedMethod    = "Exit"
)

func run(pass *analysis.Pass) (any, error) {

	for _, file := range pass.Files {

		// Is package "main"?
		if file.Name.Name != findPackageArea {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			// Find function declaration
			if fn, ok := n.(*ast.FuncDecl); ok {
				// Is function "main"?
				if fn.Name.Name == findPackageAreaFn {
					// Ispect "main":
					ast.Inspect(fn.Body, func(n ast.Node) bool {
						// Find function call
						if call, ok := n.(*ast.CallExpr); ok {
							// Is function call selector?
							if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
								// Check identificator:
								if ident, ok := sel.X.(*ast.Ident); ok {
									if ident.Name == findUsedPackage && sel.Sel.Name == findUsedMethod {
										pass.Reportf(
											sel.Pos(),
											"%s.%s usage in %s()",
											findUsedPackage,
											findUsedMethod,
											findPackageAreaFn,
										)
									}
								}
							}
						}
						return true
					})
				}
			}

			return true
		})
	}

	return nil, nil
}
