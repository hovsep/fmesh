F-Mesh is a Go-based framework inspired by Flow-Based Programming (FBP) that enables the creation of data flow networks using interconnected components, each processing signals. Components may have multiple input and output called "ports" linked via type-agnostic pipes.

# Installation

```
go get github.com/hovsep/fmesh
```

# Release naming convention

F-Mesh releases are named after 17 historical capitals of Armenia, honouring the ancient cities that played foundational roles in Armenian history. This tradition highlights the project's growth with each version, paralleling Armenia's own historical progression.


# Dependencies

Latest release has exactly one dependency used only for unit tests (which is kinda cool):

```
github.com/stretchr/testify
```

# API reference
* [FMesh](https://pkg.go.dev/github.com/hovsep/fmesh)
* [Component](https://pkg.go.dev/github.com/hovsep/fmesh/component)
* [Port](https://pkg.go.dev/github.com/hovsep/fmesh/port)
* [Signal](https://pkg.go.dev/github.com/hovsep/fmesh/signal)
* [Export](https://pkg.go.dev/github.com/hovsep/fmesh/export)
* [Common](https://pkg.go.dev/github.com/hovsep/fmesh/common)