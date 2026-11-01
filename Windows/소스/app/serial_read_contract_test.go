package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// A regression in v7.4.21 used os.File.Read for a timeout-based COM port.
// A successful idle ReadFile returning 0 bytes became io.EOF and was treated
// as a disconnection. Prevent direct File.Read calls in serial session code.
func TestCOMReadContract(t *testing.T) {
	data, err := os.ReadFile("serial_windows.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	if got := strings.Count(source, "readCOM(f, buf)"); got != 2 {
		t.Fatalf("expected 2 COM read sites, got %d", got)
	}
	fs := token.NewFileSet()
	parsed, err := parser.ParseFile(fs, "serial_windows.go", data, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(parsed, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if ok && sel.Sel.Name == "Read" {
			t.Errorf("direct File.Read in serial_windows.go may turn idle timeout into EOF")
		}
		return true
	})
}
