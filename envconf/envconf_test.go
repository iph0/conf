package envconf_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/iph0/conf"
	"github.com/iph0/conf/envconf"
)

func init() {
	os.Setenv("TEST_FOO", "bar")
	os.Setenv("TEST_MOO", "jar")
	os.Setenv("TEST_ZOO", "arr")
}

func TestLoad(t *testing.T) {
	configProc := NewProcessor()

	tConfig, err := configProc.Load(
		conf.M{
			"test": conf.M{
				"foo": conf.M{"_ref": "TEST_FOO"},
				"moo": conf.M{"_ref": "TEST_MOO"},
				"zoo": conf.M{"_ref": "TEST_ZOO"},
			},
		},

		"env:^TEST_",
	)

	if err != nil {
		t.Error(err)
		return
	}

	eConfig := conf.M{
		"test": conf.M{
			"foo": "bar",
			"moo": "jar",
			"zoo": "arr",
		},

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
	envLdr := envconf.NewLoader()

	configProc := conf.NewProcessor(
		conf.ProcessorConfig{
			Loaders: map[string]conf.Loader{
				"env": envLdr,
			},
		},
	)

	return configProc
}
