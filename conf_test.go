package conf_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/iph0/conf/v2"
)

func TestLoad(t *testing.T) {
	configProc := NewProcessor()

	tConfig, err := configProc.Load(
		"map:default",
		"map:foo",
		"map:bar",
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

			"paramDD": conf.M{
				"paramsDDA": "foo:valDDA",
				"paramsDDB": "foo:valDDB",
				"paramsDGA": "foo:valDGA",
				"paramsDDC": "foo:valDDC.G",
				"paramsDHA": "foo:valDHA",
				"paramsDGB": "foo:valDGB.H",
				"paramsDHC": "foo:valDHC",
			},

			"paramDE": "foo:bar:valDC",

			"paramDF": conf.A{
				"foo:valDFA",
				"foo:valDFB",
				"foo:foo:valDA",
			},

			"paramDG": conf.M{
				"paramsDGA": "foo:valDGA",
				"paramsDDC": "foo:valDDC.G",
				"paramsDHA": "foo:valDHA",
				"paramsDGB": "foo:valDGB.H",
				"paramsDHC": "foo:valDHC",
			},

			"paramDH": conf.M{
				"paramsDHA": "foo:valDHA",
				"paramsDGB": "foo:valDGB.H",
				"paramsDHC": "foo:valDHC",
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

			"paramDD": conf.M{
				"paramsDDA": "foo:valDDA",
				"paramsDDB": "foo:valDDB",
				"paramsDGA": "foo:valDGA",
				"paramsDDC": "foo:valDDC.G",
				"paramsDHA": "foo:valDHA",
				"paramsDGB": "foo:valDGB.H",
				"paramsDHC": "foo:valDHC",
			},

			"paramDE": "foo:bar:valDC",

			"paramDF": conf.A{
				"foo:valDFA",
				"foo:valDFB",
				"foo:foo:valDA",
			},

			"paramDG": conf.M{
				"paramsDGA": "foo:valDGA",
				"paramsDDC": "foo:valDDC.G",
				"paramsDHA": "foo:valDHA",
				"paramsDGB": "foo:valDGB.H",
				"paramsDHC": "foo:valDHC",
			},

			"paramDH": conf.M{
				"paramsDHA": "foo:valDHA",
				"paramsDGB": "foo:valDGB.H",
				"paramsDHC": "foo:valDHC",
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

		"paramS": "bar:valS",
		"paramT": "bar:valY",

		"paramU": conf.M{
			"paramUA": "foo:valUA",
			"paramUB": "foo:valUB",
			"paramUC": "foo:valUC:U",
		},

		"paramV": conf.M{
			"paramUA": "foo:valUA",
			"paramUB": "foo:valUB",
			"paramVA": "foo:valVA",
			"paramVB": "foo:valVB:V",
			"paramUC": "foo:valUC:V",
		},

		"paramW": conf.M{
			"paramUA": "foo:valUA",
			"paramUB": "foo:valUB",
			"paramVA": "foo:valVA",
			"paramUC": "foo:valUC:V",
			"paramWA": "bar:valWA",
			"paramWB": "bar:valWB",
			"paramVB": "bar:valVB:W",
		},

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

	t.Run("invalid_locator",
		func(t *testing.T) {
			_, err := configProc.Load(42)

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "configuration locator must be") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("unknown_loader",
		func(t *testing.T) {
			_, err := configProc.Load("etcd:foo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "unknown loader") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_config_type",
		func(t *testing.T) {
			_, err := configProc.Load("map:zoo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "loaded configuration must be") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_ref",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_ref")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "value of $ref directive must be") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_ref_name",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_ref_name")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "parameter name in $ref directive must be") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_ref_first_defined",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_ref_first_defined")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "\"firstDefined\" parameter in $ref directive must be") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_ref_first_defined_name",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_ref_first_defined_name")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "parameter name in \"firstDefined\" parameter must be") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_include",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_include")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "value of $include directive must be") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_include_empty",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_include_empty")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "at least one configuration locator must be sepcified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_include_locator",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_include_locator")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "configuration locator in $include directive must be") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_index",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_index")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "invalid array index") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("index_out_of_range",
		func(t *testing.T) {
			_, err := configProc.Load("map:index_out_of_range")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "index out of range") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_underlay",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_underlay")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "value of $underlay directive must be") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_underlay_empty",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_underlay_empty")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "at least one parameter name must be sepcified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_underlay_ref",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_underlay_ref")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "parameter name in $underlay directive must be") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_overlay",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_overlay")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "value of $overlay directive must be") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_overlay_empty",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_overlay_empty")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "at least one parameter name must be sepcified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_overlay_ref",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_overlay_ref")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "parameter name in $overlay directive must be") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)
}

func NewProcessor() *conf.Processor {
	var mapLdr = &mapLoader{
		m: conf.M{
			"default": conf.M{
				"paramA": "default:valA",
				"paramZ": "default:valZ",
			},

			"foo": conf.M{
				"paramA": "foo:valA",
				"paramB": "foo:valB",

				"paramD": conf.M{
					"paramDA": "foo:valDA",
					"paramDB": "foo:valDB",

					"paramDD": conf.M{
						"paramsDDA": "foo:valDDA",
						"paramsDDB": "foo:valDDB",
						"paramsDDC": "foo:valDDC.D",
						"$overlay":  "paramD.paramDG",
					},

					"paramDE": "foo:${paramD.paramDC}",

					"paramDF": conf.A{
						"foo:valDFA",
						"foo:valDFB",
						"foo:${paramD.paramDA}",
					},

					"paramDG": conf.M{
						"paramsDGA": "foo:valDGA",
						"paramsDGB": "foo:valDGB.G",
						"paramsDDC": "foo:valDDC.G",
						"$overlay":  "paramD.paramDH",
					},

					"paramDH": conf.M{
						"paramsDHA": "foo:valDHA",
						"paramsDGB": "foo:valDGB.H",
						"paramsDHC": "foo:valDHC",
					},
				},

				"paramE": conf.A{
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
					"$include": conf.A{"map:moo", "map:jar"},
				},

				"paramU": conf.M{
					"paramUA": "foo:valUA",
					"paramUB": "foo:valUB",
					"paramUC": "foo:valUC:U",
				},

				"paramV": conf.M{
					"$underlay": "paramU",
					"paramVA":   "foo:valVA",
					"paramVB":   "foo:valVB:V",
					"paramUC":   "foo:valUC:V",
				},
			},

			"bar": conf.M{
				"paramB": "bar:valB",
				"paramC": "bar:valC",

				"paramD": conf.M{
					"paramDB": "bar:valDB",
					"paramDC": "bar:valDC",
				},

				"paramE": conf.A{
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
						"firstDefined": conf.A{"paramX", "paramY"},
						"default":      "bar:valT",
					},
				},

				"paramY": "bar:valY",

				"paramW": conf.M{
					"$underlay": "paramV",
					"paramWA":   "bar:valWA",
					"paramWB":   "bar:valWB",
					"paramVB":   "bar:valVB:W",
				},
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
					"$include": "map:zoo",
				},
			},

			"zoo": conf.A{
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

			"invalid_ref_first_defined": conf.M{
				"$ref": conf.M{
					"firstDefined": 42,
					"default":      "bar:valT",
				},
			},

			"invalid_ref_first_defined_name": conf.M{
				"$ref": conf.M{
					"firstDefined": conf.A{42},
					"default":      "bar:valT",
				},
			},

			"invalid_include": conf.M{
				"paramQ": conf.M{"$include": 42},
			},

			"invalid_include_empty": conf.M{
				"paramQ": conf.M{"$include": []any{}},
			},

			"invalid_include_locator": conf.M{
				"paramQ": conf.M{"$include": []any{42}},
			},

			"invalid_index": conf.M{
				"paramQ": conf.A{"valA", "valB"},
				"paramR": conf.M{"$ref": "paramQ.paramQA"},
			},

			"index_out_of_range": conf.M{
				"paramQ": conf.A{"valA", "valB"},
				"paramR": conf.M{"$ref": "paramQ.2"},
			},

			"invalid_underlay": conf.M{
				"paramQ": conf.M{"$underlay": 42},
			},

			"invalid_underlay_empty": conf.M{
				"paramQ": conf.M{"$underlay": []any{}},
			},

			"invalid_underlay_ref": conf.M{
				"paramQ": conf.M{"$underlay": []any{42}},
			},

			"invalid_overlay": conf.M{
				"paramQ": conf.M{"$overlay": 42},
			},

			"invalid_overlay_empty": conf.M{
				"paramQ": conf.M{"$overlay": []any{}},
			},

			"invalid_overlay_ref": conf.M{
				"paramQ": conf.M{"$overlay": []any{42}},
			},
		},
	}

	configProc := conf.NewProcessor(
		conf.ProcessorConfig{
			Loaders: map[string]conf.Loader{
				"map": mapLdr,
			},
		},
	)

	return configProc
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

type mapLoader struct {
	m conf.M
}

// Load method loads configuration layer from a map.
func (l *mapLoader) Load(key string) ([]any, error) {
	return []any{l.m[key]}, nil
}
