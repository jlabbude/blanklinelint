package analyzer

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/analysis"
	"os"
	"strings"
)

var Analyzer = &analysis.Analyzer{
	Name: "blanklinelint",
	Doc:  "checks for blank lines in function bodies and lack of them between top-level declarations",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		checkTopLevelDecls(pass, file, pass.Fset)
		checkFunctionBodies(pass, file)
	}
	return nil, nil
}

func checkTopLevelDecls(pass *analysis.Pass, file *ast.File, fset *token.FileSet) {
	src := fset.File(file.Pos()).Name()
	content, err := os.ReadFile(src)
	if err != nil {
		return
	}
	lines := strings.Split(string(content), "\n")

	decls := file.Decls
	for i := 0; i < len(decls)-1; i++ {
		d1, d2 := decls[i], decls[i+1]
		d1EndLine := fset.Position(d1.End()).Line
		d2StartLine := fset.Position(d2.Pos()).Line

		hasBlank := false
		for line := d1EndLine + 1; line < d2StartLine; line++ {
			if line-1 < len(lines) && strings.TrimSpace(lines[line-1]) == "" {
				hasBlank = true
				break
			}
		}

		if !hasBlank {
			pass.Reportf(d2.Pos(), "top-level declarations should be separated by a blank line")
		}
	}
}

func checkFunctionBodies(pass *analysis.Pass, file *ast.File) {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		checkStatements(pass, fn.Body.List, pass.Fset)
	}
}

func checkStatements(pass *analysis.Pass, stmts []ast.Stmt, fset *token.FileSet) {
	for i := 0; i < len(stmts)-1; i++ {
		current, next := stmts[i], stmts[i+1]
		currentEnd := fset.Position(current.End()).Line
		nextStart := fset.Position(next.Pos()).Line

		if nextStart-currentEnd > 1 {
			pass.Reportf(next.Pos(), "unnecessary blank line between statements")
		}
		checkNestedStatements(pass, current, fset)
	}
	if len(stmts) > 0 {
		checkNestedStatements(pass, stmts[len(stmts)-1], fset)
	}
}

func checkNestedStatements(pass *analysis.Pass, stmt ast.Stmt, fset *token.FileSet) {
	var body *ast.BlockStmt
	switch s := stmt.(type) {
	case *ast.IfStmt:
		body = s.Body
	case *ast.ForStmt:
		body = s.Body
	case *ast.RangeStmt:
		body = s.Body
	case *ast.SwitchStmt:
		body = s.Body
	case *ast.TypeSwitchStmt:
		body = s.Body
	case *ast.SelectStmt:
		body = s.Body
	case *ast.BlockStmt:
		body = s
	default:
		return
	}

	if body != nil {
		checkStatements(pass, body.List, fset)
	}
}
