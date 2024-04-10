package conf_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/iph0/conf/v2"
	"github.com/iph0/conf/v2/loaders/maploader"
)

var mapLdr = maploader.NewLoader(
	conf.M{
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
				"paramDE": "foo:${paramD.paramDC}",

				"paramDF": conf.A{
					"foo:valDFA",
					"foo:valDFB",
					"foo:${paramD.paramDA}",
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
				"$include": conf.A{"map:zoo"},
			},
		},

		"zoo": conf.A{
			"zoo:valA",
			"zoo:valB",
		},

		"disabled_processing": conf.M{
			"paramA": "coo:valA",
			"paramB": "coo:${paramA}",
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

		"invalid_locator": conf.M{
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
	},
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
			Loaders: map[string]conf.Loader{
				"map": mapLdr,
			},
			DisableProcessing: true,
		},
	)

	tConfig, err := configProc.Load("map:disabled_processing")

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
			} else if strings.Index(err.Error(), "loaded configuration must be a map of type conf.M") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_ref",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_ref")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "malformed directive: $ref") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_ref_name",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_ref_name")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "reference name must be a string") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_ref_first_defined",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_ref_first_defined")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "\"firstDefined\" must be an array") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_ref_first_defined_name",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_ref_first_defined_name")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "reference name in \"firstDefined\" must be a string") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_include",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_include")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "locator list in $include directive must be an array") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_locator",
		func(t *testing.T) {
			_, err := configProc.Load("map:invalid_locator")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "locator in $include directive must be a string, but got") == -1 {
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
}

func NewProcessor() *conf.Processor {
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
