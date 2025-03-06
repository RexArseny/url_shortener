package exit_analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer check for no use of os.Exit.
var Analyzer = &analysis.Analyzer{
	Name: "exit",
	Doc:  "check for no use of os.Exit",
	Run:  run,
}

// run start the check of no use of os.Exit.
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			function, ok := node.(*ast.FuncDecl)
			if !ok {
				return true
			}
			if function.Name.Name != "main" {
				return true
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				expresion, ok := node.(*ast.ExprStmt)
				if !ok {
					return true
				}
				call, ok := expresion.X.(*ast.CallExpr)
				if !ok {
					return true
				}
				line, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if pac, ok := line.X.(*ast.Ident); ok && pac.Name == "os" && line.Sel.Name == "Exit" {
					pass.Report(analysis.Diagnostic{
						Pos:     pac.NamePos,
						End:     line.Sel.NamePos,
						Message: "use of os.Exit in main function in main package",
					})

					return false
				}

				return true
			})

			return true
		})
	}

	return nil, nil
}
