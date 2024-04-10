package envconf

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/iph0/conf/v2"
)

func init() {
	os.Setenv("TEST_FOO", "bar")
	os.Setenv("TEST_MOO", "jar")
	os.Setenv("TEST_ZOO", "arr")
}

func TestLoad(t *testing.T) {
	configProc := NewProcessor()

	tConfig, err := configProc.Load(
		"map:default",
		"env:^TEST_",
	)

	if err != nil {
		t.Error(err)
		return
	}

	eConfig := conf.M{
		"foo": "bar",
		"moo": "jar",
		"zoo": "arr",

		"TEST_FOO": "bar",
		"TEST_MOO": "jar",
		"TEST_ZOO": "arr",
	}

	if !reflect.DeepEqual(tConfig, eConfig) {
		t.Errorf("unexpected configuration returned: %#v", tConfig)
	}
}

func TestErrors(t *testing.T) {
	configProc := NewProcessor()

	t.Run("invalid_pattern",
		func(t *testing.T) {
			_, err := configProc.Load("env:^TE[ST_")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "error parsing regexp") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)
}

func NewProcessor() *conf.Processor {
	var mapLdr = &mapLoader{
		m: conf.M{
			"default": conf.M{
				"foo": conf.M{"$ref": "TEST_FOO"},
				"moo": conf.M{"$ref": "TEST_MOO"},
				"zoo": conf.M{"$ref": "TEST_ZOO"},
			},
		},
	}

	envLdr := NewLoader()

	configProc := conf.NewProcessor(
		conf.ProcessorConfig{
			Loaders: map[string]conf.Loader{
				"map": mapLdr,
				"env": envLdr,
			},
		},
	)

	return configProc
}

type mapLoader struct {
	m conf.M
}

// Load method loads configuration layer from a map.
func (l *mapLoader) Load(key string) ([]any, error) {
	return []any{l.m[key]}, nil
}
