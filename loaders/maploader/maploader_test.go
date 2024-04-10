package maploader

import (
	"reflect"
	"testing"

	"github.com/iph0/conf/v2"
)

func TestLoad(t *testing.T) {
	configProc := NewProcessor()

	tConfig, err := configProc.Load(
		"map:default",
	)

	if err != nil {
		t.Error(err)
		return
	}

	eConfig := conf.M{
		"foo": "bar",
		"moo": "jar",
		"zoo": "arr",
	}

	if !reflect.DeepEqual(tConfig, eConfig) {
		t.Errorf("unexpected configuration returned: %#v", tConfig)
	}
}

func NewProcessor() *conf.Processor {
	mapLdr := NewLoader(
		conf.M{
			"default": conf.M{
				"foo": "bar",
				"moo": "jar",
				"zoo": "arr",
			},
		},
	)

	configProc := conf.NewProcessor(
		conf.ProcessorConfig{
			Loaders: map[string]conf.Loader{
				"map": mapLdr,
			},
		},
	)

	return configProc
}
