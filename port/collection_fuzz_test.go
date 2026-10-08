package port

import (
	"strings"
	"testing"
)

// buildFuzzCollection adds one output port per distinct name and reports which
// names made it in; a duplicate must fail Add without corrupting the collection.
func buildFuzzCollection(t *testing.T, names []string) (c *Collection, known map[string]bool) {
	t.Helper()
	collection := NewCollection()
	known = make(map[string]bool, len(names))
	for _, name := range names {
		p, err := NewOutput(name)
		if err != nil {
			continue
		}
		if err := collection.Add(p); err != nil {
			if !known[name] {
				t.Fatalf("Add(%q) failed without a name conflict: %v", name, err)
			}
			continue
		}
		known[name] = true
	}
	if collection.Len() != len(known) {
		t.Fatalf("collection holds %d ports, want %d distinct names", collection.Len(), len(known))
	}
	return collection, known
}

func dedupeNames(names []string) []string {
	unique := make([]string, 0, len(names))
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if !seen[name] {
			seen[name] = true
			unique = append(unique, name)
		}
	}
	return unique
}

// FuzzCollectionNameLookup pins the forgiving-lookup contract from design.md:
// ByName returns nil for a name no port has and ByNames silently skips unknown
// names — neither may panic, whatever the names look like.
func FuzzCollectionNameLookup(f *testing.F) {
	f.Add("a,b,c", "a,b")
	f.Add("", "")
	f.Add("p1,p1,p2", "p1,missing")
	f.Add("日本,ü,", "日本,ü")
	f.Add("in1,in2", "in3")

	f.Fuzz(func(t *testing.T, portNamesCSV, lookupNamesCSV string) {
		lookupNames := strings.Split(lookupNamesCSV, ",")
		collection, known := buildFuzzCollection(t, strings.Split(portNamesCSV, ","))

		distinctKnownLookups := 0
		for _, name := range dedupeNames(lookupNames) {
			got := collection.ByName(name)
			switch {
			case known[name] && got == nil:
				t.Fatalf("ByName(%q) = nil for a port the collection holds", name)
			case known[name]:
				if got.Name() != name {
					t.Fatalf("ByName(%q) returned port %q", name, got.Name())
				}
				distinctKnownLookups++
			case got != nil:
				t.Fatalf("ByName(%q) = %q, want nil for an unknown name", name, got.Name())
			}
		}

		// Duplicated lookup names collapse to one match apiece: unknown names are
		// skipped, known ones kept exactly once. (An earlier ByNames dropped every
		// name after a duplicate — found by this fuzzer, seed "ü,"/",,ü".)
		subset := collection.ByNames(lookupNames...)
		if subset.Len() != distinctKnownLookups {
			t.Fatalf("ByNames returned %d ports, want %d (unknown names must be skipped, known ones kept once)",
				subset.Len(), distinctKnownLookups)
		}
		for p := range subset.All() {
			if !known[p.Name()] {
				t.Fatalf("ByNames returned port %q the collection does not hold", p.Name())
			}
		}
	})
}
