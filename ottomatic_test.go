package ottomatic

import (
	"testing"

	"github.com/robertkrimen/otto"
)

type OttoFn func(otto.FunctionCall) otto.Value

type ObjectWithMethods struct {
	Name   string             `otto:"name"`
	Sum    func(int, int) int `otto:"sum"`
	Inner  *InnerObject       `otto:"inner"`
	SkipMe bool               `otto:"-"`
	NoTag  int
}

type InnerObject struct {
	Value int `otto:"value"`
}

func TestDeepGet(t *testing.T) {
	o := otto.New()
	if _, err := o.Run(`parent = { child: {grandchild: "hello"}};`); err != nil {
		t.Fatal(err)
	}

	v, err := DeepGet("parent.child.grandchild", o)
	if err != nil {
		t.Error(err)
	}

	if str, err := v.ToString(); err != nil {
		t.Error(err)
	} else if str != "hello" {
		t.Errorf("Expected \"hello\", got %q", str)
	}

	for _, tt := range []struct {
		selector  string
		undefined string
	}{
		{"parent.nosuchchild", "nosuchchild"},
		{"parent.nosuchchild.nosuchgrandchild", "nosuchchild"},
		{"parent.child.nosuchgrandchild", "nosuchgrandchild"},
	} {
		v, err = DeepGet(tt.selector, o)
		if err == nil {
			t.Error("Expected error for %q", tt.selector)
		}
		ee, ok := err.(ErrUndefined)
		if !ok {
			t.Errorf("Expected undefined error, got %s", err)
		}
		if string(ee) != tt.undefined {
			t.Errorf("Expected %q, got %q", tt.undefined, string(ee))
		}
	}
}

func TestRegister(t *testing.T) {
	o := otto.New()
	Register("hello", "world", o)

	res, err := DeepGet("hello", o)
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

	// Canary to make sure our def satisfies the interface.
	var _ OttoFn = fn
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
		Name:   "astro",
		Sum:    func(a, b int) int { return a + b },
		Inner:  &InnerObject{Value: 42},
		SkipMe: true,
		NoTag:  24,
	}

	if err := Register("top", owm, o); err != nil {
		t.Fatal(err)
	}

	if res, err := DeepGet("top.SkipMe", o); err == nil {
		t.Fatalf("Fetched undefined object. Should get undefined. Got error %s", err)
	} else if res != otto.UndefinedValue() {
		t.Errorf("Expected undefined object, got %v", res)
	} else if undef := err.(ErrUndefined); string(undef) != "SkipMe" {
		t.Errorf("Expected undefined object to be SkipMe, got %v", res)
	}

	// Test that we can access objects from within the runtime.

	for js, expect := range map[string]int{
		"top.inner.value": 42,
		"top.NoTag":       24,
	} {
		res, err := DeepGet(js, o)
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

	if res, err := DeepGet("top.name", o); err != nil {
		t.Errorf("No top.name: %s", err)
	} else if res == otto.UndefinedValue() {
		t.Error("undefined: top.name")
	} else if s, err := res.ToString(); s != "astro" {
		t.Errorf("Expected astro, got %s (%s)", s, err)
	}

	// Test a simple script
	o.Set("NoTag", 1)
	o.Set("innervalue", 2)
	script := `var myval = top.NoTag + top.inner.value`
	if out, err := o.Run(script); err != nil {
		t.Fatalf("failed to run script %q: %s", script, err)
	} else {
		t.Logf("Output: %v", out)
	}

	if res, err := DeepGet("myval", o); err != nil {
		t.Fatal(err)
	} else if res == otto.UndefinedValue() {
		t.Error("'myval' is undefined")
	} else if ival, err := res.ToInteger(); ival != 66 {
		t.Errorf("Expected 66, got %d (%v)", ival, err)
	}

	// Test that we can execute a function.
	script = `myval = top.sum(10, 21);`
	if out, err := o.Run(script); err != nil {
		t.Fatalf("failed to run script %q: %v", script, err)
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
