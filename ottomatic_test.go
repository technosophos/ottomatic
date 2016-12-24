package ottomatic

import (
	"testing"

	"github.com/robertkrimen/otto"
)

type ObjectWithMethods struct {
	Name   string                    `otto:"name"`
	Sum    func(a, b int) otto.Value `otto:"sum"`
	Inner  InnerObject               `otto:"inner"`
	SkipMe bool                      `otto:"-"`
	NoTag  int
}

type InnerObject struct {
	Value int `otto:"value"`
}

func TestRegister(t *testing.T) {
	o := otto.New()
	owm := &ObjectWithMethods{
		Name: "astro",
		Sum: func(a, b int) otto.Value {
			ret, err := o.ToValue(a + b)
			if err != nil {
				panic(err)
			}
			return ret
		},
		Inner:  InnerObject{Value: 42},
		SkipMe: true,
		NoTag:  24,
	}

	if err := Register("top", owm, o); err != nil {
		t.Fatal(err)
	}
}
