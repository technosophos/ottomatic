# Ottomatic: Annotation-based automatic binding for Otto

This library provides a simple way of binding a Go struct to a JavaScript
definition for the [Otto](https://github.com/robertkrimen/otto) JavaScript runtime.

## Usage

Annotate objects with `otto:` tags:

```go
type Object struct {
	Name   string                    `otto:"name"`
	Sum    func(a, b int) otto.Value `otto:"sum"`
	Inner  InnerObject               `otto:"inner"`
	SkipMe bool                      `otto:"-"`
	NoTag  int
}
```

Register your object with an Otto runtime:

```go
rt := otto.New()
myobj := &Object{}

if err := ottomatic.Register("myobj", myobj, ottoruntime); err != nil {
  // Handle error
}
```

Your JavaScript runtime will now have access to `myobj`, `myobj.name`,
`myobj.NoTag` and so on.

Some kinds cannot be automatically bound to Otto, including UnsafePtr
and Chan. But the commonly used types, including slices, maps, and
structs, can be bound.
