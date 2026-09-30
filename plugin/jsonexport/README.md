# jsonexport

A mesh plugin that exports the mesh structure as JSON: components, ports, pipes, descriptions and
metadata. Use it to feed other tools, diff two versions of a mesh, or snapshot it in a test.

```go
import "github.com/hovsep/fmesh/plugin/jsonexport"

exporter := jsonexport.New()
fm, err := fmesh.New("pricing", fmesh.WithPlugins(exporter))
// ... add components and pipes ...

data, err := exporter.Export() // indented JSON
```

Output:

```json
{
  "name": "pricing",
  "meta": {"env": "prod"},
  "components": [
    {"name": "dst", "inputs": [{"name": "in"}], "outputs": []},
    {"name": "src", "description": "emits prices", "inputs": [], "outputs": [{"name": "out"}]}
  ],
  "pipes": [
    {"from": {"component": "src", "port": "out"}, "to": {"component": "dst", "port": "in"}}
  ]
}
```

- The order is fixed: components and ports by name, pipes in wiring order. The same mesh always
  exports the same bytes.
- `description` and `meta` are left out when empty. `components`, `inputs`, `outputs` and `pipes`
  are always present.
- The types `Mesh`, `Component`, `Port`, `Pipe` and `Endpoint` describe the shape; unmarshal into
  `jsonexport.Mesh` to read an export back.
- One plugin instance serves one mesh. `Export` before the plugin is attached returns
  `ErrNotAttached`.
