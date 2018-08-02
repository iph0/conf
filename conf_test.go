package conf_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/iph0/conf"
)

type mapLoader struct {
	layers map[string]interface{}
}

func TestLoad(t *testing.T) {
	configProc := NewProcessor()

	tConfig, err := configProc.Load(
		map[string]interface{}{
			"paramA": "default:valA",
			"paramZ": "default:valZ",
		},

		"test:foo",
		"test:bar",
	)

	if err != nil {
		t.Error(err)
		return
	}

	eConfig := map[string]interface{}{
		"paramA": "foo:valA",
		"paramB": "bar:valB",
		"paramC": "bar:valC",

		"paramD": map[string]interface{}{
			"paramDA": "foo:valDA",
			"paramDB": "bar:valDB",
			"paramDC": "bar:valDC",
			"paramDE": "foo:bar:valDC",

			"paramDF": []interface{}{
				"foo:valDFA",
				"foo:valDFB",
				"foo:foo:valDA",
			},
		},

		"paramE": []interface{}{
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

		"paramM": map[string]interface{}{
			"paramDA": "foo:valDA",
			"paramDB": "bar:valDB",
			"paramDC": "bar:valDC",
			"paramDE": "foo:bar:valDC",

			"paramDF": []interface{}{
				"foo:valDFA",
				"foo:valDFB",
				"foo:foo:valDA",
			},
		},

		"paramN": map[string]interface{}{
			"paramNA": "foo:valNA",
			"paramNB": "foo:valNB",

			"paramNC": map[string]interface{}{
				"paramNCA": "foo:valNCA",
				"paramNCB": "bar:valNCB",
				"paramNCC": "bar:valNCC",
				"paramNCD": "bar:foo:valNCA",
				"paramNCE": "foo:valNB",
			},
		},

		"paramO": map[string]interface{}{
			"paramOA": "moo:valOA",
			"paramOB": "jar:valOB",
			"paramOC": "jar:valOC",

			"paramOD": map[string]interface{}{
				"paramODA": "moo:valODA",
				"paramODB": "jar:valODB",
				"paramODC": "jar:valODC",
				"paramODD": "jar:bar:valNCB",
			},

			"paramOE": []interface{}{
				"zoo:valA",
				"zoo:valB",
			},
		},

		"paramP": map[string]interface{}{
			"paramODA": "moo:valODA",
			"paramODB": "jar:valODB",
			"paramODC": "jar:valODC",
			"paramODD": "jar:bar:valNCB",
		},

		"paramZ": "default:valZ",
	}

	if !reflect.DeepEqual(tConfig, eConfig) {
		t.Errorf("unexpected configuration returned: %#v", tConfig)
	}
}

func TestDisableProcessing(t *testing.T) {
	configProc := conf.NewProcessor(
		conf.ProcessorConfig{
			DisableProcessing: true,
		},
	)

	tConfig, err := configProc.Load(
		map[string]interface{}{
			"paramA": "coo:valA",
			"paramB": "coo:${paramA}",
		},
	)

	if err != nil {
		t.Error(err)
		return
	}

	eConfig := map[string]interface{}{
		"paramA": "coo:valA",
		"paramB": "coo:${paramA}",
	}

	if !reflect.DeepEqual(tConfig, eConfig) {
		t.Errorf("unexpected configuration returned: %#v", tConfig)
	}
}

func TestPanic(t *testing.T) {
	t.Run("no_locators",
		func(t *testing.T) {
			defer func() {
				err := recover()
				errStr := fmt.Sprintf("%v", err)

				if err == nil {
					t.Error("no error happened")
				} else if strings.Index(errStr, "no configuration locators") == -1 {
					t.Error("other error happened:", err)
				}
			}()

			configProc := NewProcessor()
			configProc.Load()
		},
	)
}

func TestErrors(t *testing.T) {
	configProc := NewProcessor()

	t.Run("empty_locator",
		func(t *testing.T) {
			_, err := configProc.Load("")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "empty configuration locator") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_locator",
		func(t *testing.T) {
			_, err := configProc.Load(42)

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "locator has invalid type") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("missing_loader",
		func(t *testing.T) {
			_, err := configProc.Load("foo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "missing loader name") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("loader_not_found",
		func(t *testing.T) {
			_, err := configProc.Load("etcd:foo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "loader not found") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_config_type",
		func(t *testing.T) {
			_, err := configProc.Load("test:zoo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "has invalid type") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_var",
		func(t *testing.T) {
			_, err := configProc.Load("test:invalid_var")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "invalid _var directive") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_include",
		func(t *testing.T) {
			_, err := configProc.Load("test:invalid_include")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "invalid _include directive") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_index",
		func(t *testing.T) {
			_, err := configProc.Load("test:invalid_index")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "invalid slice index") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("index_out_of_range",
		func(t *testing.T) {
			_, err := configProc.Load("test:index_out_of_range")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "index out of range") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)
}

func NewProcessor() *conf.Processor {
	mapLdr := NewLoader()

	configProc := conf.NewProcessor(
		conf.ProcessorConfig{
			Loaders: map[string]conf.Loader{
				"test": mapLdr,
			},
		},
	)

	return configProc
}

func NewLoader() conf.Loader {
	return &mapLoader{
		map[string]interface{}{
			"foo": map[string]interface{}{
				"paramA": "foo:valA",
				"paramB": "foo:valB",

				"paramD": map[string]interface{}{
					"paramDA": "foo:valDA",
					"paramDB": "foo:valDB",
					"paramDE": "foo:${.paramDC}",

					"paramDF": []interface{}{
						"foo:valDFA",
						"foo:valDFB",
						"foo:${..paramDA}",
					},
				},

				"paramE": []interface{}{
					"foo:valEA",
					"foo:valEB",
				},

				"paramF": "foo:${paramB}",
				"paramH": "foo:${paramE.0}",
				"paramJ": "foo:${paramI}",
				"paramL": "foo:$${paramD.paramDE}:${}:$${paramD.paramDA}",

				"paramN": map[string]interface{}{
					"paramNA": "foo:valNA",
					"paramNB": "foo:valNB",

					"paramNC": map[string]interface{}{
						"paramNCA": "foo:valNCA",
						"paramNCB": "foo:valNCB",
						"paramNCE": map[string]interface{}{"_var": "..paramNB"},
					},
				},

				"paramO": map[string]interface{}{
					"_include": []interface{}{"test:moo", "test:jar"},
				},
			},

			"bar": map[string]interface{}{
				"paramB": "bar:valB",
				"paramC": "bar:valC",

				"paramD": map[string]interface{}{
					"paramDB": "bar:valDB",
					"paramDC": "bar:valDC",
				},

				"paramE": []interface{}{
					"bar:valEA",
					"bar:valEB",
				},

				"paramG": "bar:${paramD.paramDA}",
				"paramI": "bar:${paramH}",
				"paramK": "bar:${paramD.paramDF.1}:${paramD.paramDE}",
				"paramM": map[string]interface{}{"_var": "paramD"},

				"paramN": map[string]interface{}{
					"paramNC": map[string]interface{}{
						"paramNCB": "bar:valNCB",
						"paramNCC": "bar:valNCC",
						"paramNCD": "bar:${paramN.paramNC.paramNCA}",
					},
				},

				"paramP": map[string]interface{}{"_var": "paramO.paramOD"},
			},

			"moo": map[string]interface{}{
				"paramOA": "moo:valOA",
				"paramOB": "moo:valOB",

				"paramOD": map[string]interface{}{
					"paramODA": "moo:valODA",
					"paramODB": "moo:valODB",
				},
			},

			"jar": map[string]interface{}{
				"paramOB": "jar:valOB",
				"paramOC": "jar:valOC",

				"paramOD": map[string]interface{}{
					"paramODB": "jar:valODB",
					"paramODC": "jar:valODC",
					"paramODD": "jar:${paramN.paramNC.paramNCB}",
				},

				"paramOE": map[string]interface{}{
					"_include": []interface{}{"test:zoo"},
				},
			},

			"zoo": []interface{}{
				"zoo:valA",
				"zoo:valB",
			},

			"invalid_var": map[string]interface{}{
				"paramQ": map[string]interface{}{"_var": 42},
			},

			"invalid_include": map[string]interface{}{
				"paramQ": map[string]interface{}{"_include": 42},
			},

			"invalid_index": map[string]interface{}{
				"paramQ": []interface{}{"valA", "valB"},
				"paramR": map[string]interface{}{"_var": "paramQ.paramQA"},
			},

			"index_out_of_range": map[string]interface{}{
				"paramQ": []interface{}{"valA", "valB"},
				"paramR": map[string]interface{}{"_var": "paramQ.2"},
			},
		},
	}
}

func (p *mapLoader) Load(loc *conf.Locator) (interface{}, error) {
	key := loc.BareLocator
	layer, _ := p.layers[key]

	return layer, nil
}
