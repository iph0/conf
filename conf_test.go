package conf_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/iph0/conf/v2"
)

type mapLoader struct {
	layers conf.M
}

func TestLoad(t *testing.T) {
	configProc := NewProcessor()

	tConfig, err := configProc.Load(
		conf.M{
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

	eConfig := conf.M{
		"paramA": "foo:valA",
		"paramB": "bar:valB",
		"paramC": "bar:valC",

		"paramD": conf.M{
			"paramDA": "foo:valDA",
			"paramDB": "bar:valDB",
			"paramDC": "bar:valDC",
			"paramDE": "foo:bar:valDC",

			"paramDF": conf.S{
				"foo:valDFA",
				"foo:valDFB",
				"foo:foo:valDA",
			},
		},

		"paramE": conf.S{
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

			"paramDF": conf.S{
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

			"paramOE": conf.S{
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

		"paramS": "bar:valS",
		"paramT": "bar:valY",
		"paramY": "bar:valY",
		"paramZ": "default:valZ",
	}

	if !reflect.DeepEqual(tConfig, eConfig) {
		t.Errorf("unexpected configuration returned: %+v is not equal to %+v",
			tConfig, eConfig)
	}
}

func TestDisableProcessing(t *testing.T) {
	configProc := conf.NewProcessor(
		conf.ProcessorConfig{
			DisableProcessing: true,
		},
	)

	tConfig, err := configProc.Load(
		conf.M{
			"paramA": "coo:valA",
			"paramB": "coo:${paramA}",
		},
	)

	if err != nil {
		t.Error(err)
		return
	}

	eConfig := conf.M{
		"paramA": "coo:valA",
		"paramB": "coo:${paramA}",
	}

	if !reflect.DeepEqual(tConfig, eConfig) {
		t.Errorf("unexpected configuration returned: %+v is not equal to %+v",
			tConfig, eConfig)
	}
}

func TestDecode(t *testing.T) {
	type testConfig struct {
		ParamA string `conf:"test_paramA"`
		ParamB int    `conf:"test_paramB"`
		ParamC []string
		ParamD map[string]bool
	}

	configRaw := conf.M{
		"test_paramA": "foo:val",
		"test_paramB": 1234,
		"paramC":      []string{"moo:val1", "moo:val2"},
		"paramD": map[string]bool{
			"zoo": true,
			"arr": false,
		},
	}

	var tConfig testConfig
	conf.Decode(configRaw, &tConfig)

	eConfig := testConfig{
		ParamA: "foo:val",
		ParamB: 1234,
		ParamC: []string{"moo:val1", "moo:val2"},
		ParamD: map[string]bool{
			"zoo": true,
			"arr": false,
		},
	}

	if !reflect.DeepEqual(tConfig, eConfig) {
		t.Errorf("unexpected configuration returned: %+v is not equal to %+v",
			tConfig, eConfig)
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
			} else if strings.Index(err.Error(), "invalid type of configuration locator") == -1 {
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

	t.Run("invalid_ref",
		func(t *testing.T) {
			_, err := configProc.Load("test:invalid_ref")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "invalid type of $ref directive") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_ref_name",
		func(t *testing.T) {
			_, err := configProc.Load("test:invalid_ref_name")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "invalid type of reference name") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_ref_first_of",
		func(t *testing.T) {
			_, err := configProc.Load("test:invalid_ref_first_of")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "invalid type of \"firstOf\" field") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_ref_first_of_name",
		func(t *testing.T) {
			_, err := configProc.Load("test:invalid_ref_first_of_name")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "invalid type of reference name") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_include",
		func(t *testing.T) {
			_, err := configProc.Load("test:invalid_include")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "invalid type of $include directive") == -1 {
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
		conf.M{
			"foo": conf.M{
				"paramA": "foo:valA",
				"paramB": "foo:valB",

				"paramD": conf.M{
					"paramDA": "foo:valDA",
					"paramDB": "foo:valDB",
					"paramDE": "foo:${paramD.paramDC}",

					"paramDF": conf.S{
						"foo:valDFA",
						"foo:valDFB",
						"foo:${paramD.paramDA}",
					},
				},

				"paramE": conf.S{
					"foo:valEA",
					"foo:valEB",
				},

				"paramF": "foo:${paramB}",
				"paramH": "foo:${paramE.0}",
				"paramJ": "foo:${paramI}",
				"paramL": "foo:$${paramD.paramDE}:${}:$${paramD.paramDA}",

				"paramN": conf.M{
					"paramNA": "foo:valNA",
					"paramNB": "foo:valNB",

					"paramNC": conf.M{
						"paramNCA": "foo:valNCA",
						"paramNCB": "foo:valNCB",
						"paramNCE": conf.M{"$ref": "paramN.paramNB"},
					},
				},

				"paramO": conf.M{
					"$include": conf.S{"test:moo", "test:jar"},
				},
			},

			"bar": conf.M{
				"paramB": "bar:valB",
				"paramC": "bar:valC",

				"paramD": conf.M{
					"paramDB": "bar:valDB",
					"paramDC": "bar:valDC",
				},

				"paramE": conf.S{
					"bar:valEA",
					"bar:valEB",
				},

				"paramG": "bar:${paramD.paramDA}",
				"paramI": "bar:${paramH}",
				"paramK": "bar:${paramD.paramDF.1}:${paramD.paramDE}",
				"paramM": conf.M{"$ref": "paramD"},

				"paramN": conf.M{
					"paramNC": conf.M{
						"paramNCB": "bar:valNCB",
						"paramNCC": "bar:valNCC",
						"paramNCD": "bar:${paramN.paramNC.paramNCA}",
					},
				},

				"paramP": conf.M{"$ref": "paramO.paramOD"},

				"paramS": conf.M{
					"$ref": conf.M{
						"name":    "paramX",
						"default": "bar:valS",
					},
				},

				"paramT": conf.M{
					"$ref": conf.M{
						"firstOf": conf.S{"paramX", "paramY"},
						"default": "bar:valT",
					},
				},

				"paramY": "bar:valY",
			},

			"moo": conf.M{
				"paramOA": "moo:valOA",
				"paramOB": "moo:valOB",

				"paramOD": conf.M{
					"paramODA": "moo:valODA",
					"paramODB": "moo:valODB",
				},
			},

			"jar": conf.M{
				"paramOB": "jar:valOB",
				"paramOC": "jar:valOC",

				"paramOD": conf.M{
					"paramODB": "jar:valODB",
					"paramODC": "jar:valODC",
					"paramODD": "jar:${paramN.paramNC.paramNCB}",
				},

				"paramOE": conf.M{
					"$include": conf.S{"test:zoo"},
				},
			},

			"zoo": conf.S{
				"zoo:valA",
				"zoo:valB",
			},

			"invalid_ref": conf.M{
				"paramQ": conf.M{"$ref": 42},
			},

			"invalid_ref_name": conf.M{
				"$ref": conf.M{
					"name":    42,
					"default": "foo",
				},
			},

			"invalid_ref_first_of": conf.M{
				"$ref": conf.M{
					"firstOf": 42,
					"default": "bar:valT",
				},
			},

			"invalid_ref_first_of_name": conf.M{
				"$ref": conf.M{
					"firstOf": conf.S{42},
					"default": "bar:valT",
				},
			},

			"invalid_include": conf.M{
				"paramQ": conf.M{"$include": 42},
			},

			"invalid_index": conf.M{
				"paramQ": conf.S{"valA", "valB"},
				"paramR": conf.M{"$ref": "paramQ.paramQA"},
			},

			"index_out_of_range": conf.M{
				"paramQ": conf.S{"valA", "valB"},
				"paramR": conf.M{"$ref": "paramQ.2"},
			},
		},
	}
}

func (p *mapLoader) Load(loc *conf.Locator) (any, error) {
	key := loc.Value
	layer, _ := p.layers[key]

	return layer, nil
}

func ExampleDecode() {
	type DBConfig struct {
		Host     string `conf:"server_host"`
		Port     int    `conf:"server_port"`
		DBName   string
		Username string
		Password string
	}

	configRaw := conf.M{
		"server_host": "stat.mydb.com",
		"server_port": 1234,
		"dbname":      "stat",
		"username":    "stat_writer",
		"password":    "some_pass",
	}

	var config DBConfig
	conf.Decode(configRaw, &config)

	fmt.Printf("%v", config)

	// Output: {stat.mydb.com 1234 stat stat_writer some_pass}
}
