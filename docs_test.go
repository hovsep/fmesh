package fmesh_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Documentation drifts silently: a rename leaves the wiki telling users to call
// something that no longer exists, and nothing fails until someone copies it.
//
// The README is compile-checked by Example in example_test.go. Wiki snippets are
// fragments with elisions, so compiling them is guesswork — but every qualified
// reference to this project's API in them must name a real exported symbol, which
// is exactly the drift a rename causes and has no false positives.
//
// Not covered: argument counts, method calls (the receiver type is unknowable in
// a fragment), and prose outside Go blocks.

// docPackages maps the import qualifier used in docs to its directory.
var docPackages = map[string]string{
	"fmesh":     ".",
	"component": "component",
	"port":      "port",
	"signal":    "signal",
	"meta":      "meta",
	"cycle":     "cycle",
	"plugin":    "plugin",
}

var (
	goBlockRe   = regexp.MustCompile("(?s)```go\n(.*?)\n```")
	qualifiedRe = regexp.MustCompile(`\b(fmesh|component|port|signal|meta|cycle|plugin)\.([A-Z]\w*)`)
	lineComment = regexp.MustCompile(`//.*`)
)

// exportedSymbols returns a package's exported top-level names plus its method
// names. Methods count because docs shadow package names with variables —
// `port.Signals()` is a *Port called port. A deleted name is gone from both sets,
// so a rename is still caught.
func exportedSymbols(t *testing.T, dir string) map[string]bool {
	t.Helper()

	symbols := make(map[string]bool)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	fset := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.SkipObjectResolution)
		require.NoError(t, err)

		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name.IsExported() {
					symbols[d.Name.Name] = true
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name.IsExported() {
							symbols[s.Name.Name] = true
						}
					case *ast.ValueSpec:
						for _, ident := range s.Names {
							if ident.IsExported() {
								symbols[ident.Name] = true
							}
						}
					}
				}
			}
		}
	}
	return symbols
}

func TestDocs_ReferenceOnlyExistingAPI(t *testing.T) {
	symbols := make(map[string]map[string]bool, len(docPackages))
	for qualifier, dir := range docPackages {
		symbols[qualifier] = exportedSymbols(t, dir)
	}

	files, err := filepath.Glob("docs/wiki/*.md")
	require.NoError(t, err)
	files = append(files, "README.md", "CHANGELOG.md", "CONTRIBUTING.md")

	type reference struct {
		file, qualifier, symbol string
	}
	var missing []reference

	for _, file := range files {
		content, err := os.ReadFile(file)
		require.NoError(t, err)

		for _, block := range goBlockRe.FindAllStringSubmatch(string(content), -1) {
			// Comments discuss the API in prose ("AddLabel is gone"), which is not a
			// reference to it.
			code := lineComment.ReplaceAllString(block[1], "")
			for _, ref := range qualifiedRe.FindAllStringSubmatch(code, -1) {
				qualifier, symbol := ref[1], ref[2]
				if !symbols[qualifier][symbol] {
					missing = append(missing, reference{file, qualifier, symbol})
				}
			}
		}
	}

	if len(missing) == 0 {
		return
	}

	seen := make(map[string]bool)
	lines := make([]string, 0, len(missing))
	for _, m := range missing {
		line := m.file + ": " + m.qualifier + "." + m.symbol
		if !seen[line] {
			seen[line] = true
			lines = append(lines, line)
		}
	}
	sort.Strings(lines)
	t.Fatalf("documentation references %d symbol(s) that do not exist:\n  %s",
		len(lines), strings.Join(lines, "\n  "))
}
