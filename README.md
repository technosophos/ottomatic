# Ottomatic: Annotation-based automatic binding for Otto

This library provides a simple way of binding a Go struct to a JavaScript
definition for the [Otto](https://github.com/robertkrimen/otto) JavaScript runtime.

## Installation

The easisest way to install is with [Glide](http://glide.sh):

```console
$ glide init     # if you haven't already
$ glide get github.com/technosophos/ottomatic
```

## Usage

```go
import (
  "github.com/robertkrimen/otto"
  "github.com/technosophos/ottomatic"
)
```

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

## The `otto` annotation

The `otto` annotation follows the general tag convention used in Go:

```
FIELD TYPE `otto:"NAME,PARAM,PARAM"`
```

- `NAME` is the name by which the JavaScript runtime will be able to
  access the object.
  - The special name `-` indicates that this field should be ignored,
    and not exposed inside of the JavaScript runtime.
- `PARAM` is always optional, and is list-like. The following parameters
  are defined:
  - `alias=ALTNAME` (example: `otto:"kubernetes,alias=k8s"`) instructs
    ottomatic to register this object again, but under the given
    alternative name. If the field is a pointer, both handles will
    point to the same object. In any other case, each handle will have
    its own target value. _More than one alias may be specified._
  - `omitempty`: Reserved for future use.
  - `returns`, `returns=`, `throws`, and `throws=` reserved for future use.

All unknown params are _silently ignored_.

If no annotation is specified and the field is exportable (i.e. the Go
field name starts with an uppercase letter), the field will be exported
to the JavaScript runtime using its Go name.
