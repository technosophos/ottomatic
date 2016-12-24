package ottomatic

import (
	"errors"
	"reflect"
	"strings"

	"github.com/robertkrimen/otto"
)

var ErrUnsupportedKind = errors.New("unsupported kind")

// OttoTagName is the name of the struct field that this scans for otto definitions.
//
// This is a var, rather than a constant, because it is feasible (though perhaps not
// the best idea) to use JSON tags to extract this information.
var OttoTagName = "otto"

type ottoTag struct {
	name string
	omit bool
}

// Register registers v into the JavaScript object o, with the name n.
//
// This will attempt to bind v in its entirety. If v is a struct, this will bind
// it according to the 'ott:' tags on fields.
func Register(n string, v interface{}, o *otto.Otto) error {
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
		o.Set(n, s)
		// Handle struct scanning
		t := val.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			ot := gettag(&f)
			if !ot.omit {
				name := strings.Join([]string{n, ot.name}, ".")
				Register(name, val.Field(i).Interface(), o)
			}
		}
		return nil
	default:
		o.Set(n, v)
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
