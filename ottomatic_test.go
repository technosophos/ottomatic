package ottomatic

import (
	"testing"

	"github.com/robertkrimen/otto"
)

type OttoFn func(otto.FunctionCall) otto.Value

type ObjectWithMethods struct {
	Name   string       `otto:"name"`
	Sum    OttoFn       `otto:"sum"`
	Inner  *InnerObject `otto:"inner"`
	SkipMe bool         `otto:"-"`
	NoTag  int
}

type InnerObject struct {
	Value int `otto:"value"`
}

func TestRegister(t *testing.T) {
	o := otto.New()
	Register("hello", "world", o)

	res, err := o.Get("hello")
	if err != nil {
		t.Fatal(err)
	}

	v, err := res.ToString()
	if err != nil {
		t.Fatal(err)
	}
	if v != "world" {
		t.Errorf("Expected hello=\"world\", got hello=%q", v)
	}
}

func TestRegister_Func(t *testing.T) {
	o := otto.New()
	fn := func(a otto.FunctionCall) otto.Value { ret, _ := o.ToValue("world"); return ret }
	Register("hello", fn, o)

	res, err := o.Run("hello();")
	if err != nil {
		t.Fatal(err)
	}

	v, err := res.ToString()
	if err != nil {
		t.Fatal(err)
	}
	if v != "world" {
		t.Errorf("Expected hello=\"world\", got hello=%q", v)
	}
}

func TestRegister_Struct(t *testing.T) {
	o := otto.New()
	owm := &ObjectWithMethods{
		Name: "astro",
		Sum: func(args otto.FunctionCall) otto.Value {
			a, _ := args.Argument(0).ToInteger()
			b, _ := args.Argument(1).ToInteger()
			ret, err := o.ToValue(a + b)
			if err != nil {
				panic(err)
			}
			return ret
		},
		Inner:  &InnerObject{Value: 42},
		SkipMe: true,
		NoTag:  24,
	}

	if err := Register("top", owm, o); err != nil {
		t.Fatal(err)
	}

	if res, err := o.Get("top.SkipMe"); err != nil {
		t.Fatal("Fetched undefined object. Should get undefined. Got error %s", err)
	} else if res != otto.UndefinedValue() {
		t.Errorf("Expected undefined object, got %v", res)
	}

	// Test that we can access objects from within the runtime.

	for js, expect := range map[string]int{
		"top.inner.value": 42,
		"top.NoTag":       24,
	} {
		res, err := o.Get(js)
		if err != nil {
			t.Fatal(err)
		}

		ival, err := res.ToInteger()
		if err != nil {
			t.Fatalf("%s: error converting val %v to int: %s", js, res, err)
		}
		if int(ival) != expect {
			t.Errorf("%s: Expected %d, got %d", js, expect, ival)
		}
	}

	if res, err := o.Get("top.name"); err != nil {
		t.Errorf("No top.name: %s", err)
	} else if res == otto.UndefinedValue() {
		t.Error("undefined: top.name")
	} else if s, err := res.ToString(); s != "astro" {
		t.Errorf("Expected astro, got %s (%s)", s, err)
	}

	// Test that we can execute a function.

	script := `myval = top.sum(10, 21);`
	if out, err := o.Run(script); err != nil {
		t.Fatalf("failed to run script %q: %s", script, err)
	} else {
		t.Logf("Output: %v", out)
	}

	if res, err := o.Get("myval"); err != nil {
		t.Fatal(err)
	} else if res == otto.UndefinedValue() {
		t.Error("'myval' is undefined")
	} else if ival, err := res.ToInteger(); ival != 31 {
		t.Errorf("Expected 31, got %s (%s)", ival, err)
	}
}
