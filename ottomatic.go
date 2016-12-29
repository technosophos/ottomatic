package ottomatic

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/robertkrimen/otto"
)

// OttoTagName is the name of the struct field that this scans for otto definitions.
//
// This is a var, rather than a constant, because it is feasible (though perhaps not
// the best idea) to use JSON tags to extract this information.
var OttoTagName = "otto"

type ottoTag struct {
	name string
	omit bool
}

// ErrUnsupportedKind indicates that a given kind is not supported by the registry.
var ErrUnsupportedKind = errors.New("unsupported kind")

// Undefined is the error version of otto.Value == Undefined.
//
// There are a number of situations in which an undefined value in
// JavaScript should be treated as a Go error. This error captures that intent.
type ErrUndefined string

func (e ErrUndefined) Error() string {
	return fmt.Sprintf("undefined value for %q", string(e))
}

// ObjectSetter can set a JavaScript value on an object.
type ObjectSetter interface {
	Set(name string, val interface{}) error
}

// ObjectGetter can get a value from a JavaScript object.
type ObjectGetter interface {
	Get(name string) (otto.Value, error)
}

// Register registers v into the JavaScript object o, with the name n.
//
// This will attempt to bind v in its entirety. If v is a struct, this will bind
// it according to the 'ott:' tags on fields.
func Register(n string, v interface{}, o *otto.Otto) error {
	// Here, Otto is an ObjectSetter, so we can bind to the root namespace by
	// binding directly to the Otto runtime.
	return RegisterTo(n, v, o, o)
}

// RegisterTo registers n to v on the object given in obj.
//
// This behaves like Register, with the additional stipulation that it binds
// to the passed-in object instead of to the root of the Otto runtime's namespace.
func RegisterTo(n string, v interface{}, o *otto.Otto, obj ObjectSetter) error {
	val := reflect.Indirect(reflect.ValueOf(v))
	switch val.Kind() {
	// TODO: are reflect.Interface, reflect.Ptr, and reflect.Uintptr okay?
	// TODO: can Complex64/128 be represented by Otto
	// TODO: is there any processing we need to do on maps, slices, and arrays?
	case reflect.UnsafePointer, reflect.Chan, reflect.Invalid:
		return ErrUnsupportedKind
	case reflect.Struct:
		s, err := o.Object(n + " = {}")
		if err != nil {
			return err
		}
		// Handle struct scanning
		t := val.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			ot := gettag(&f)
			if !ot.omit {
				RegisterTo(ot.name, val.Field(i).Interface(), o, s)
			}
		}
		obj.Set(n, s)
		return nil
	default:
		obj.Set(n, v)
		return nil
	}
}

func gettag(field *reflect.StructField) ottoTag {
	t := field.Tag.Get(OttoTagName)
	if len(t) == 0 {
		return ottoTag{name: field.Name}
	}

	// We preserve the convention used by JSON, YAML, and other tags so that
	// an evil library user can change OttoTagName to somethiing and get rational
	// output from this.
	data := strings.Split(t, ",")
	n := data[0]
	if n == "-" {
		return ottoTag{name: field.Name, omit: true}
	}
	return ottoTag{name: n}
}

// DeepGet gets a value from an object. The key may reference objects in dotted notation.
//
// The standard Get methods on Otto can only fetch the value of a direct child.
// This makes it possible to get the descrendants of an object. The key may
// use JavaScript dotted notation ('parent.child.grandchild') to specify the
// target.
//
// This function does not support fetching values whose keys contain dots.
// Dots are used exclusively as namespace separators. To fetch keys with dots
// in their names, use one of Otto's built-in Get functions.
//
// DeepGet does not provide a way to reference an array index.
//
// DeepGet returns an ErrUndefined if any object in the requested key
// does not exist (i.e. is undefined) in the JavaScript runtime. It also
// returns the otto.Value in that case (which will always be a JavaScript
// undefined).
func DeepGet(key string, o ObjectGetter) (otto.Value, error) {
	keys := strings.Split(key, ".")
	var v otto.Value
	var err error
	obj := o
	for _, k := range keys {
		v, err = obj.Get(k)
		if err != nil {
			return v, err
		}
		if v.IsUndefined() {
			return v, ErrUndefined(k)
		}
		obj = v.Object()
		if obj == nil {
			return v, errors.New("nil object")
		}
	}
	return v, err
}
