// This file keeps the documentation honest about the API.
//
// Nothing else does. The README's quick start is compiled by Example in
// example_test.go, but the wiki is 4,000 lines of Markdown that no compiler ever
// reads, so a rename leaves it telling users to call something that no longer
// exists and nothing fails until someone copies the snippet. Both checks here
// exist because that had already happened.
//
//   - TestDocs_ReferenceOnlyExistingAPI — every qualified reference in a Go block
//     (component.X, signal.X, …) must name a real exported symbol.
//   - TestDocs_NoRemovedMethodNames — catches what the first one cannot: calls on
//     a variable, like c.Labels().AddLabel(...), whose receiver type a fragment
//     does not reveal. It works by refusing a list of names that were removed.
//   - TestDocs_MethodCallsExistSomewhere — the positive version of the same idea:
//     every `.SomeExported(` call in a fenced Go block must name a method or
//     function defined somewhere in this module, or sit on a small allowlist of
//     external names (stdlib, the profiler module) that snippets legitimately use.
//     The denylist test stays because it also scans Go source comments.
//
// Deliberately not checked: argument counts, and prose outside Go blocks — the
// wiki discusses removed API on purpose when explaining a migration.
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

// docPackages maps the import qualifier used in docs to its directory.
var docPackages = map[string]string{
	"fmesh":     ".",
	"component": "component",
	"port":      "port",
	"signal":    "signal",
	"meta":      "meta",
	"cycle":     "cycle",
	"profiler":  "plugin/profiler",
	"autowire":  "plugin/autowire",
}

var (
	goBlockRe   = regexp.MustCompile("(?s)```go\n(.*?)\n```")
	qualifiedRe = regexp.MustCompile(`\b(fmesh|component|port|signal|meta|cycle|profiler|autowire)\.([A-Z]\w*)`)
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
	pluginReadmes, err := filepath.Glob("plugin/*/README.md")
	require.NoError(t, err)
	files = append(files, pluginReadmes...)
	files = append(files, "plugin/README.md")

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

// methodCallRe matches a method call on something other than a package
// qualifier: `.Foo(`, where Foo is exported.
var methodCallRe = regexp.MustCompile(`\.([A-Z]\w*)\(`)

// TestDocs_NoRemovedMethodNames catches the case
// TestDocs_ReferenceOnlyExistingAPI is documented as unable to see: a call on a
// variable rather than a package, like `c.Labels().AddLabel("k", "v")`. The
// receiver's type is unknowable in a fragment, so the general check skips these
// — which is how three snippets went on calling AddLabel and SetLabels for a
// release after both were deleted.
//
// The narrow version is exact: a method name that exists nowhere in this module
// cannot be a valid call on any of its types. Only names that once existed and
// were removed need listing; add to it whenever public API is renamed away.
func TestDocs_NoRemovedMethodNames(t *testing.T) {
	removed := []string{
		"AddLabel", "AddLabels", "SetLabels", "ClearLabels", "RemoveLabels",
		"AddScalar", "AddScalars", "SetScalars", "ClearScalars", "RemoveScalars",
		"PayloadOrNil", "PayloadOrDefault",
		"HasChainableErr", "ChainableErr",
		"GetName",
		"SetLabelOnEach", "SetScalarOnEach",
		"SumScalar", "MinScalar", "MaxScalar", "AvgScalar",
		"NewIndexedInputGroup", "NewIndexedOutputGroup",
	}
	banned := make(map[string]bool, len(removed))
	for _, name := range removed {
		banned[name] = true
	}

	// A name is only stale if nothing in the module defines it any more.
	for qualifier, dir := range docPackages {
		for symbol := range exportedSymbols(t, dir) {
			if banned[symbol] {
				t.Fatalf("%s.%s is in the removed list but still exists — drop it from the list",
					qualifier, symbol)
			}
		}
	}

	files, err := filepath.Glob("docs/wiki/*.md")
	require.NoError(t, err)
	files = append(files, "README.md", "CONTRIBUTING.md")
	pluginReadmes, err := filepath.Glob("plugin/*/README.md")
	require.NoError(t, err)
	files = append(files, pluginReadmes...)
	files = append(files, "plugin/README.md")

	// Go sources too. A removed method cannot appear in compiling code, so any
	// hit here is necessarily a comment — which is exactly where one survived
	// undetected in port/collection.go's ForEach example.
	goFiles, err := filepath.Glob("*/*.go")
	require.NoError(t, err)
	rootGo, err := filepath.Glob("*.go")
	require.NoError(t, err)
	files = append(files, append(goFiles, rootGo...)...)

	var stale []string
	for _, file := range files {
		if file == "docs_test.go" {
			continue // names every removed method by definition
		}
		content, err := os.ReadFile(file)
		require.NoError(t, err)
		blocks := []string{}
		if strings.HasSuffix(file, ".go") {
			blocks = append(blocks, string(content))
		} else {
			for _, b := range goBlockRe.FindAllStringSubmatch(string(content), -1) {
				// Markdown prose discusses removed API deliberately; only fenced
				// Go blocks claim to be usable code.
				blocks = append(blocks, lineComment.ReplaceAllString(b[1], ""))
			}
		}
		for _, code := range blocks {
			for _, call := range methodCallRe.FindAllStringSubmatch(code, -1) {
				if banned[call[1]] {
					stale = append(stale, file+": ."+call[1]+"()")
				}
			}
		}
	}

	if len(stale) > 0 {
		sort.Strings(stale)
		t.Fatalf("documentation calls %d removed method(s):\n  %s",
			len(stale), strings.Join(stale, "\n  "))
	}
}

// moduleDefinedNames returns every exported method and function name defined in
// any non-test .go file in this module, subpackages and internal/ included —
// promoted methods (collection.Keyed, meta.store) are defined there.
func moduleDefinedNames(t *testing.T) map[string]bool {
	t.Helper()

	names := make(map[string]bool)
	fset := token.NewFileSet()
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if name := d.Name(); name != "." && (strings.HasPrefix(name, ".") || name == "testdata") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok && fd.Name.IsExported() {
				names[fd.Name.Name] = true
			}
		}
		return nil
	})
	require.NoError(t, err)
	return names
}

// docsExternalNames is the allowlist for TestDocs_MethodCallsExistSomewhere:
// names that exist outside this module but appear as calls in snippets. Keep it
// minimal — every entry is a name the check can no longer question.
var docsExternalNames = map[string]bool{
	// stdlib
	"Background": true, "WithCancel": true, // context
	"Errorf": true, "Printf": true, "Println": true, "Sprintf": true, // fmt / log
	"HasPrefix": true, "ToLower": true, "ToUpper": true, // strings
	"Now": true, "Since": true, "Format": true, // time
	"Is":    true,              // errors
	"Itoa":  true,              // strconv
	"Clone": true, "Min": true, // maps / slices
	"Load": true,                                    // sync/atomic
	"Exit": true, "Getenv": true, "WriteFile": true, // os
	"NewRequestWithContext": true, // net/http
	// the DOT exporter (fmesh-graphviz) and the profiler (fmesh-profiler) live in
	// their own modules; the wiki shows their entry points
	"Export": true, "ExportWithCycles": true, "NewDotExporter": true,
	"Report": true,
	// placeholder interface in a teaching snippet (sink.Publish, 602)
	"Publish": true,
}

// TestDocs_MethodCallsExistSomewhere is the positive complement of the denylist:
// a `.SomeExported(` call in a fenced Go block whose name no file in this module
// defines (and that is not on the external allowlist) cannot be a valid call on
// any of this module's types. Unlike the denylist it also catches renames nobody
// remembered to list.
func TestDocs_MethodCallsExistSomewhere(t *testing.T) {
	known := moduleDefinedNames(t)

	files, err := filepath.Glob("docs/wiki/*.md")
	require.NoError(t, err)
	// CHANGELOG.md is deliberately absent: its fenced blocks show removed API as
	// "before" migration examples.
	files = append(files, "README.md", "CONTRIBUTING.md")
	pluginReadmes, err := filepath.Glob("plugin/*/README.md")
	require.NoError(t, err)
	files = append(files, pluginReadmes...)
	files = append(files, "plugin/README.md")

	var unknown []string
	seen := make(map[string]bool)
	for _, file := range files {
		content, err := os.ReadFile(file)
		require.NoError(t, err)
		for _, block := range goBlockRe.FindAllStringSubmatch(string(content), -1) {
			code := lineComment.ReplaceAllString(block[1], "")
			for _, call := range methodCallRe.FindAllStringSubmatch(code, -1) {
				name := call[1]
				if known[name] || docsExternalNames[name] {
					continue
				}
				line := file + ": ." + name + "()"
				if !seen[line] {
					seen[line] = true
					unknown = append(unknown, line)
				}
			}
		}
	}

	if len(unknown) > 0 {
		sort.Strings(unknown)
		t.Fatalf("documentation calls %d method(s) defined nowhere in this module:\n  %s",
			len(unknown), strings.Join(unknown, "\n  "))
	}
}
