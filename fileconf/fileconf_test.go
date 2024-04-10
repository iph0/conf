package fileconf

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/iph0/conf/v2"
)

func TestLoad(t *testing.T) {
	configProc, err := NewProcessor()

	if err != nil {
		t.Error(err)
		return
	}

	tConfig, err := configProc.Load(
		"map:default",
		"file:foo.yml",
		"file:bar.json",
	)

	if err != nil {
		t.Error(err)
		return
	}

	eConfig := conf.M{
		"paramA": "foo:valA",
		"paramB": "bar:valB",
		"paramC": "bar:valC",

		"paramD": conf.M{
			"paramDA": "foo:valDA",
			"paramDB": "bar:valDB",
			"paramDC": "bar:valDC",
			"paramDE": "foo:bar:valDC",

			"paramDF": conf.A{
				"foo:valDFA",
				"foo:valDFB",
				"foo:foo:valDA",
			},
		},

		"paramE": conf.A{
			"bar:valEA",
			"bar:valEB",
		},

		"paramF": "foo:bar:valB",
		"paramG": "bar:foo:valDA",
		"paramH": "foo:bar:valEA",
		"paramI": "bar:foo:bar:valEA",
		"paramJ": "foo:bar:foo:bar:valEA",
		"paramK": "bar:foo:valDFB:foo:bar:valDC",
		"paramL": "foo:${paramD.paramDE}:${}:${paramD.paramDA}",

		"paramM": conf.M{
			"paramDA": "foo:valDA",
			"paramDB": "bar:valDB",
			"paramDC": "bar:valDC",
			"paramDE": "foo:bar:valDC",

			"paramDF": conf.A{
				"foo:valDFA",
				"foo:valDFB",
				"foo:foo:valDA",
			},
		},

		"paramN": conf.M{
			"paramNA": "foo:valNA",
			"paramNB": "foo:valNB",

			"paramNC": conf.M{
				"paramNCA": "foo:valNCA",
				"paramNCB": "bar:valNCB",
				"paramNCC": "bar:valNCC",
				"paramNCD": "bar:foo:valNCA",
				"paramNCE": "foo:valNB",
			},
		},

		"paramO": conf.M{
			"paramOA": "moo:valOA",
			"paramOB": "jar:valOB",
			"paramOC": "jar:valOC",

			"paramOD": conf.M{
				"paramODA": "moo:valODA",
				"paramODB": "jar:valODB",
				"paramODC": "jar:valODC",
				"paramODD": "jar:bar:valNCB",
			},

			"paramOE": conf.A{
				"zoo:valA",
				"zoo:valB",
			},
		},

		"paramP": conf.M{
			"paramODA": "moo:valODA",
			"paramODB": "jar:valODB",
			"paramODC": "jar:valODC",
			"paramODD": "jar:bar:valNCB",
		},

		"paramZ": "default:valZ",
	}

	if !reflect.DeepEqual(tConfig, eConfig) {
		t.Errorf("unexpected configuration returned: %+v is not equal to %+v",
			tConfig, eConfig)
	}
}

func TestPanic(t *testing.T) {
	t.Run("no_directories",
		func(t *testing.T) {
			defer func() {
				err := recover()
				errStr := fmt.Sprintf("%v", err)

				if err == nil {
					t.Error("no error happened")
				} else if strings.Index(errStr, "no directories specified") == -1 {
					t.Error("other error happened:", errStr)
				}
			}()

			NewLoader()
		},
	)
}

func TestErrors(t *testing.T) {
	configProc, err := NewProcessor()

	if err != nil {
		t.Error(err)
		return
	}

	t.Run("file_extension_not_specified",
		func(t *testing.T) {
			_, err := configProc.Load("file:coo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "file extension not specified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("unknown_file_extension",
		func(t *testing.T) {
			_, err := configProc.Load("file:mar.html")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "unknown file extension") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_pattern",
		func(t *testing.T) {
			_, err := configProc.Load("file:f[oo.yml")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "syntax error in pattern") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)
}

func NewProcessor() (*conf.Processor, error) {
	mapLdr := &mapLoader{
		m: conf.M{
			"default": conf.M{
				"paramA": "default:valA",
				"paramZ": "default:valZ",
			},
		},
	}

	fileLdr := NewLoader("fileconf_test/etc")

	configProc := conf.NewProcessor(
		conf.ProcessorConfig{
			Loaders: map[string]conf.Loader{
				"map":  mapLdr,
				"file": fileLdr,
			},
		},
	)

	return configProc, nil
}

type mapLoader struct {
	m conf.M
}

// Load method loads configuration layer from a map.
func (l *mapLoader) Load(loc *conf.Locator) (any, error) {
	return l.m[loc.Value], nil
}
