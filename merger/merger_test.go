package merger_test

import (
	"reflect"
	"testing"

	"github.com/iph0/conf/merger"
)

type Foo struct {
	FooA string
	FooB float64
	FooC []int
	FooD map[string]interface{}
	FooE *Bar
	Err  error

	fooF int
}

type Bar struct {
	BarA string
	BarB float64
	BarC map[string]interface{}

	barD int
}

func TestMerge(t *testing.T) {
	a := map[string]interface{}{
		"foo": Foo{
			FooA: "Hello!",
			FooC: []int{1, 2, 3},
			FooD: map[string]interface{}{
				"moo": 67,
				"jar": "Yahoo!",
			},
			FooE: &Bar{
				BarB: 45.65,
				BarC: map[string]interface{}{
					"goo": "Ola!",
					"yar": 77.7,
				},
			},
		},
		"bar": "Yahoo!",
		"coo": 15,
		"zar": nil,
	}

	b := map[string]interface{}{
		"foo": Foo{
			FooB: 42,
			FooD: map[string]interface{}{
				"moo": 1024,
				"zoo": "Rrrr!",
			},
			FooE: &Bar{
				BarA: "Zap!",
				BarC: map[string]interface{}{
					"zoo": "Grrrr!",
					"var": "Yes!",
				},
			},
		},
		"coo": "Hello!",
		"mar": 15,
		"zar": nil,
	}

	cTest := merger.Merge(a, b)

	cExp := map[string]interface{}{
		"foo": Foo{
			FooA: "Hello!",
			FooB: 42,
			FooC: []int{1, 2, 3},
			FooD: map[string]interface{}{
				"moo": 1024,
				"jar": "Yahoo!",
				"zoo": "Rrrr!",
			},
			FooE: &Bar{
				BarA: "Zap!",
				BarB: 45.65,
				BarC: map[string]interface{}{
					"goo": "Ola!",
					"yar": 77.7,
					"zoo": "Grrrr!",
					"var": "Yes!",
				},
			},
		},
		"bar": "Yahoo!",
		"coo": "Hello!",
		"mar": 15,
	}

	if !reflect.DeepEqual(cTest, cExp) {
		t.Errorf("unexpected configuration returned: %#v", cTest)
	}
}
