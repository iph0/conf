package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/iph0/conf/merger"
)

// Foo is a struct for example
type Foo struct {
	FooA string
	FooB float64
	FooC []int
	FooD map[string]interface{}
	FooE *Bar

	fooF int
}

// Bar is a struct for example
type Bar struct {
	BarA string
	BarB float64
	BarC map[string]interface{}

	barD int
}

func init() {
	spew.Config.Indent = "  "
}

func main() {
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

	c := merger.Merge(a, b)

	spew.Dump(c)
}
