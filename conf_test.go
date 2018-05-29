package conf_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/iph0/conf"
)

type testDriver struct {
	sections map[string]map[string]interface{}
}

func TestLoad(t *testing.T) {
	loader := getLoader()

	tConfig, err := loader.Load(
		"test:dirs",
		"test:db",
		"test:unknown",

		map[string]interface{}{
			"myapp": map[string]interface{}{
				"db": map[string]interface{}{
					"connectors": map[string]interface{}{
						"stat": map[string]interface{}{
							"host": "localhost",
							"port": 4321,
						},
					},
				},
			},
		},
	)

	if err != nil {
		t.Error("failed to load configuration:", err)
		return
	}

	eConfig := map[string]interface{}{
		"myapp": map[string]interface{}{
			"mediaFormats": []string{"images", "audio", "video"},
			"pageTitles":   []string{"images", "audio", "video"},
			"metadata":     "foo:${moo.jar}:bar",

			"dirs": map[string]interface{}{
				"rootDir":      "/myapp",
				"templatesDir": "/myapp/templates",
				"sessionsDir":  "/myapp/sessions",
				"mediaDirs": []interface{}{
					"/myapp/media/images",
					"/myapp/media/audio",
					"/myapp/media/video",
				},
			},

			"db": map[string]interface{}{
				"connectors": map[string]interface{}{
					"stat": map[string]interface{}{
						"host":     "localhost",
						"port":     4321,
						"dbname":   "stat",
						"username": "stat_writer",
						"password": "stat_writer_pass",
					},

					"metrics": map[string]interface{}{
						"host":     "metrics.mydb.com",
						"port":     4321,
						"dbname":   "metrics",
						"username": "metrics_writer",
						"password": "metrics_writer_pass",
					},
				},
			},

			"servers": map[string]interface{}{
				"alpha": map[string]interface{}{
					"ip": "10.0.0.1",
					"dc": "foodc",
				},

				"beta": map[string]interface{}{
					"ip": "10.0.0.2",
					"dc": "foodc",
				},
			},
		},
	}

	if !reflect.DeepEqual(tConfig, eConfig) {
		t.Errorf("unexpected configuration returned: %#v", tConfig)
	}
}

func TestErrors(t *testing.T) {
	loader := getLoader()

	t.Run("empty_pattern",
		func(t *testing.T) {
			_, err := loader.Load("")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "empty pattern specified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("no_driver",
		func(t *testing.T) {
			_, err := loader.Load("foo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "missing driver name in pattern") == -1 {
				t.Error("other error happened:", err)
			}

			_, err = loader.Load(":foo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "missing driver name in pattern") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("unknown_driver",
		func(t *testing.T) {
			_, err := loader.Load("redis:foo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "unknown pattern specified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("driver_error",
		func(t *testing.T) {
			_, err := loader.Load("test:invalid")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "something wrong") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)
}

func TestPanic(t *testing.T) {
	t.Run("no_drivers",
		func(t *testing.T) {
			defer func() {
				err := recover()
				errStr := fmt.Sprintf("%v", err)

				if err == nil {
					t.Error("no error happened")
				} else if strings.Index(errStr, "no drivers specified") == -1 {
					t.Error("other error happened:", errStr)
				}
			}()

			conf.NewLoader()
		},
	)

	t.Run("invalid_type",
		func(t *testing.T) {
			defer func() {
				err := recover()
				errStr := fmt.Sprintf("%v", err)

				if err == nil {
					t.Error("no error happened")
				} else if strings.Index(errStr, "is invalid type") == -1 {
					t.Error("other error happened:", errStr)
				}
			}()

			loader := getLoader()

			_, err := loader.Load(42)

			if err != nil {
				t.Error(err)
			}
		},
	)
}

func getLoader() *conf.Loader {
	var driver = &testDriver{
		map[string]map[string]interface{}{
			"dirs": {
				"myapp": map[string]interface{}{
					"mediaFormats": []string{"images", "audio", "video"},
					"pageTitles":   map[string]interface{}{"@var": ".mediaFormats"},
					"metadata":     "foo:$${moo.jar}:bar",

					"dirs": map[string]interface{}{
						"rootDir":      "/myapp",
						"templatesDir": "${myapp.dirs.rootDir}/templates",
						"sessionsDir":  "${.rootDir}/sessions",
						"mediaDirs": []interface{}{
							"${..rootDir}/media/${myapp.mediaFormats.0}",
							"${..rootDir}/media/${myapp.mediaFormats.1}",
							"${..rootDir}/media/${myapp.mediaFormats.2}",
						},
					},

					"servers": map[string]interface{}{"@include": "test:servers"},
				},
			},

			"db": {
				"myapp": map[string]interface{}{
					"db": map[string]interface{}{
						"connectors": map[string]interface{}{
							"stat": map[string]interface{}{
								"host":     "stat.mydb.com",
								"port":     1234,
								"dbname":   "stat",
								"username": "stat_writer",
								"password": "stat_writer_pass",
							},

							"metrics": map[string]interface{}{
								"host":     "metrics.mydb.com",
								"port":     4321,
								"dbname":   "metrics",
								"username": "metrics_writer",
								"password": "metrics_writer_pass",
							},
						},
					},
				},
			},

			"servers": map[string]interface{}{
				"alpha": map[string]interface{}{
					"ip": "10.0.0.1",
					"dc": "foodc",
				},

				"beta": map[string]interface{}{
					"ip": "10.0.0.2",
					"dc": "foodc",
				},
			},
		},
	}

	return conf.NewLoader(driver)
}

func (d *testDriver) Name() string {
	return "test"
}

func (d *testDriver) Load(key string) (interface{}, error) {
	if key == "invalid" {
		return nil, errors.New("something wrong")
	}

	config, ok := d.sections[key]

	if !ok {
		return nil, nil
	}

	return config, nil
}
