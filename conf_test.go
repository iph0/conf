package conf_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/iph0/conf"
)

type updatesNotifier struct{}

func init() {
	conf.RegisterProvider("map",
		func() (conf.Provider, error) {
			provider := conf.NewMapProvider(
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
								"paramNCE": map[string]interface{}{"@var": "..paramNB"},
							},
						},

						"paramO": map[string]interface{}{
							"@include": []interface{}{"map:moo", "map:jar"},
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
						"paramM": map[string]interface{}{"@var": "paramD"},

						"paramN": map[string]interface{}{
							"paramNC": map[string]interface{}{
								"paramNCB": "bar:valNCB",
								"paramNCC": "bar:valNCC",
								"paramNCD": "bar:${paramN.paramNC.paramNCA}",
							},
						},

						"paramP": map[string]interface{}{"@var": "paramO.paramOD"},
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
							"@include": []interface{}{"map:zoo"},
						},
					},

					"zoo": []interface{}{
						"zoo:valA",
						"zoo:valB",
					},

					"invalidVar": map[string]interface{}{
						"paramQ": map[string]interface{}{"@var": 42},
					},

					"invalidInclude": map[string]interface{}{
						"paramQ": map[string]interface{}{"@include": 42},
					},

					"invalidLocator": map[string]interface{}{
						"paramQ": map[string]interface{}{
							"@include": []interface{}{42},
						},
					},

					"invalidIndexA": map[string]interface{}{
						"paramQ": []interface{}{"valA", "valB"},
						"paramR": map[string]interface{}{"@var": "paramQ.paramQA"},
					},

					"invalidIndexB": map[string]interface{}{
						"paramQ": []interface{}{"valA", "valB"},
						"paramR": map[string]interface{}{"@var": "paramQ.2"},
					},
				},
			)

			return provider, nil
		},
	)
}

func TestLoad(t *testing.T) {
	loader, err := conf.NewLoader(
		conf.LoaderConfig{
			Locators: []string{
				"map:foo",
				"map:bar",
			},

			Watch: &updatesNotifier{},
		},
	)

	if err != nil {
		t.Error(err)
		return
	}

	tConfig, err := loader.Load()

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
	}

	if !reflect.DeepEqual(tConfig, eConfig) {
		t.Errorf("unexpected configuration returned: %#v", tConfig)
	}

	loader.Close()
}

func TestErrors(t *testing.T) {
	t.Run("no_locators",
		func(t *testing.T) {
			_, err := conf.NewLoader(
				conf.LoaderConfig{
					Locators: []string{},
				},
			)

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "no configuration locators") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("empty_locator",
		func(t *testing.T) {
			_, err := conf.NewLoader(
				conf.LoaderConfig{
					Locators: []string{""},
				},
			)

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "empty configuration locator") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("missing_provider",
		func(t *testing.T) {
			_, err := conf.NewLoader(
				conf.LoaderConfig{
					Locators: []string{"foo"},
				},
			)

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "missing provider name") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("provider_not_found",
		func(t *testing.T) {
			_, err := conf.NewLoader(
				conf.LoaderConfig{
					Locators: []string{"file:foo"},
				},
			)

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "provider not found") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("layer_not_found",
		func(t *testing.T) {
			loader, err := conf.NewLoader(
				conf.LoaderConfig{
					Locators: []string{"map:unknown"},
				},
			)

			if err != nil {
				t.Error(err)
				return
			}

			_, err = loader.Load()

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "layer not found") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_config_type",
		func(t *testing.T) {
			loader, err := conf.NewLoader(
				conf.LoaderConfig{
					Locators: []string{"map:zoo"},
				},
			)

			if err != nil {
				t.Error(err)
				return
			}

			_, err = loader.Load()

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "has invalid type") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_var",
		func(t *testing.T) {
			loader, err := conf.NewLoader(
				conf.LoaderConfig{
					Locators: []string{"map:invalidVar"},
				},
			)

			if err != nil {
				t.Error(err)
				return
			}

			_, err = loader.Load()

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "invalid @var directive") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_include",
		func(t *testing.T) {
			loader, err := conf.NewLoader(
				conf.LoaderConfig{
					Locators: []string{"map:invalidInclude"},
				},
			)

			if err != nil {
				t.Error(err)
				return
			}

			_, err = loader.Load()

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "invalid @include directive") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_locator_type",
		func(t *testing.T) {
			loader, err := conf.NewLoader(
				conf.LoaderConfig{
					Locators: []string{"map:invalidLocator"},
				},
			)

			if err != nil {
				t.Error(err)
				return
			}

			_, err = loader.Load()

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "locator has invalid type") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_index",
		func(t *testing.T) {
			loader, err := conf.NewLoader(
				conf.LoaderConfig{
					Locators: []string{"map:invalidIndexA"},
				},
			)

			if err != nil {
				t.Error(err)
				return
			}

			_, err = loader.Load()

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "invalid slice index") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("index_out_of_range",
		func(t *testing.T) {
			loader, err := conf.NewLoader(
				conf.LoaderConfig{
					Locators: []string{"map:invalidIndexB"},
				},
			)

			if err != nil {
				t.Error(err)
				return
			}

			_, err = loader.Load()

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "index out of range") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)
}

func (n *updatesNotifier) Notify(provider string) {}
