package merger_test

import (
	"fmt"
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

	tc := merger.Merge(a, b)

	ec := map[string]interface{}{
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

	if !reflect.DeepEqual(tc, ec) {
		t.Errorf("unexpected configuration returned: %#v", tc)
	}
}

func ExampleMerge() {
	type Connector struct {
		Host     string
		Port     int
		Username string
		Password string
		DBName   string
	}

	defaultConnrs := map[string]Connector{
		"stat": Connector{
			Port:     1234,
			Username: "stat_writer",
			DBName:   "stat",
		},

		"messages": Connector{
			Host:     "messages.mydb.com",
			Port:     5678,
			Username: "moo",
			Password: "moo_pass",
			DBName:   "messages",
		},
	}

	connrs := map[string]Connector{
		"stat": Connector{
			Host:     "stat.mydb.com",
			Username: "foo",
			Password: "foo_pass",
		},

		"metrics": Connector{
			Host:     "metrics.mydb.com",
			Port:     4321,
			Username: "bar",
			Password: "bar_pass",
			DBName:   "metrics",
		},
	}

	connrs = merger.Merge(defaultConnrs, connrs).(map[string]Connector)

	for name, connr := range connrs {
		fmt.Printf("%s: %v\n", name, connr)
	}

	// Unordered output:
	// stat: {stat.mydb.com 1234 foo foo_pass stat}
	// messages: {messages.mydb.com 5678 moo moo_pass messages}
	// metrics: {metrics.mydb.com 4321 bar bar_pass metrics}
}
