package collection

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type named string

func (n named) Name() string { return string(n) }

func newSlice(items ...int) *Slice[int] {
	s := &Slice[int]{}
	AppendItems(s, items...)
	return s
}

func TestSlice_ReadSurface(t *testing.T) {
	tests := []struct {
		name      string
		slice     *Slice[int]
		wantLen   int
		wantFirst int
		wantLast  int
	}{
		{
			name:    "empty slice returns zero values",
			slice:   newSlice(),
			wantLen: 0,
		},
		{
			name:      "items keep insertion order",
			slice:     newSlice(3, 1, 2),
			wantLen:   3,
			wantFirst: 3,
			wantLast:  2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantLen, tt.slice.Len())
			assert.Equal(t, tt.wantLen == 0, tt.slice.IsEmpty())
			assert.Equal(t, tt.wantFirst, tt.slice.First())
			assert.Equal(t, tt.wantLast, tt.slice.Last())
		})
	}
}

func TestSlice_All(t *testing.T) {
	s := newSlice(3, 1, 2)

	var visited []int
	for item := range s.All() {
		visited = append(visited, item)
		if item == 1 {
			break
		}
	}
	assert.Equal(t, []int{3, 1}, visited, "All yields in insertion order and honors early stop")

	all := slices.Collect(s.All())
	all[0] = 99
	assert.Equal(t, 3, s.First(), "a collected slice is independent of the group")
}

func TestSlice_Predicates(t *testing.T) {
	even := func(i int) bool { return i%2 == 0 }
	tests := []struct {
		name      string
		slice     *Slice[int]
		wantAny   bool
		wantEvery bool
		wantCount int
		wantFind  int
	}{
		{
			name:      "empty slice is vacuously every",
			slice:     newSlice(),
			wantAny:   false,
			wantEvery: true,
			wantCount: 0,
			wantFind:  0,
		},
		{
			name:      "mixed items",
			slice:     newSlice(1, 2, 3, 4),
			wantAny:   true,
			wantEvery: false,
			wantCount: 2,
			wantFind:  2,
		},
		{
			name:      "all match",
			slice:     newSlice(2, 4),
			wantAny:   true,
			wantEvery: true,
			wantCount: 2,
			wantFind:  2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantAny, tt.slice.Any(even))
			assert.Equal(t, tt.wantEvery, tt.slice.Every(even))
			assert.Equal(t, tt.wantCount, tt.slice.Count(even))
			assert.Equal(t, tt.wantFind, tt.slice.Find(even))
		})
	}
}

func TestSlice_RawAndReplace(t *testing.T) {
	s := newSlice(1, 2)
	assert.Equal(t, []int{1, 2}, Items(s))
	SetItems(s, []int{5})
	assert.Equal(t, []int{5}, slices.Collect(s.All()))
}

func TestKeyed_Add(t *testing.T) {
	tests := []struct {
		name        string
		existing    []named
		add         []named
		wantErr     string
		wantOrdered []named
	}{
		{
			name:        "items are traversed in name order",
			add:         []named{"c", "a", "b"},
			wantOrdered: []named{"a", "b", "c"},
		},
		{
			name:        "new items merge into the existing order",
			existing:    []named{"b", "d"},
			add:         []named{"e", "a", "c"},
			wantOrdered: []named{"a", "b", "c", "d", "e"},
		},
		{
			name:        "a duplicate within the batch adds nothing",
			add:         []named{"a", "b", "a"},
			wantErr:     `widget "a" already exists`,
			wantOrdered: nil,
		},
		{
			name:        "a name already in the collection adds nothing",
			existing:    []named{"b"},
			add:         []named{"a", "b"},
			wantErr:     `widget "b" already exists`,
			wantOrdered: []named{"b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := NewKeyed[named]("widget")
			require.NoError(t, k.Add(tt.existing...))
			err := k.Add(tt.add...)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantOrdered, slices.Collect(k.All()))
			assert.Equal(t, len(tt.wantOrdered), k.Len())
		})
	}
}

func TestKeyed_Lookups(t *testing.T) {
	k := NewKeyed[named]("widget")
	require.NoError(t, k.Add("b", "a"))

	assert.Equal(t, named("a"), k.ByName("a"))
	assert.Equal(t, named(""), k.ByName("missing"))
	assert.Equal(t, named("a"), k.First(), "first in name order, not insertion order")
	assert.Equal(t, 2, k.Len())
	assert.False(t, k.IsEmpty())
}

func TestKeyed_AllIsIndependent(t *testing.T) {
	k := NewKeyed[named]("widget")
	require.NoError(t, k.Add("a"))
	slices.Collect(k.All())[0] = "z"
	assert.Equal(t, named("a"), k.ByName("a"))
	assert.Equal(t, []named{"a"}, slices.Collect(k.All()))
}

func TestKeyed_Remove(t *testing.T) {
	k := NewKeyed[named]("widget")
	require.NoError(t, k.Add("a", "b", "c"))

	k.Remove("b", "missing")
	assert.Equal(t, []named{"a", "c"}, slices.Collect(k.All()))
}

func TestKeyed_Traversal(t *testing.T) {
	k := NewKeyed[named]("widget")
	require.NoError(t, k.Add("c", "a", "b"))

	var visited []named
	for item := range k.All() {
		visited = append(visited, item)
		if item == "b" {
			break
		}
	}
	assert.Equal(t, []named{"a", "b"}, visited, "All yields in name order and honors early stop")

	visited = nil
	for item := range k.All() {
		visited = append(visited, item)
		if item == "a" {
			require.NoError(t, k.Add("d"))
			k.Remove("c")
		}
	}
	assert.Equal(t, []named{"a", "b", "c"}, visited, "a loop body may change the collection; the walk keeps its snapshot")
	assert.Equal(t, []named{"a", "b", "d"}, slices.Collect(k.All()))
}

func TestKeyed_Predicates(t *testing.T) {
	isA := func(n named) bool { return n == "a" }

	empty := NewKeyed[named]("widget")
	assert.True(t, empty.Every(isA), "vacuous truth on empty")
	assert.False(t, empty.Any(isA))
	assert.Equal(t, named(""), empty.Find(isA))
	assert.Equal(t, named(""), empty.First())

	k := NewKeyed[named]("widget")
	require.NoError(t, k.Add("a", "b"))
	assert.True(t, k.Any(isA))
	assert.False(t, k.Every(isA))
	assert.Equal(t, 1, k.Count(isA))
	assert.Equal(t, named("a"), k.Find(isA))
}
