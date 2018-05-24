package conf_test

import (
	"fmt"
	"net/url"
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
		"test://dirs",
		"test://db",
		"test://unknown.yml",

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
		t.Error("loading of configuration failed")
	}

	eConfig := map[string]interface{}{
		"myapp": map[string]interface{}{
			"mediaFormats": []string{"images", "audio", "video"},
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
		},
	}

	if !reflect.DeepEqual(tConfig, eConfig) {
		t.Errorf("unexpected configuration returned: %#v", tConfig)
	}
}

func TestErrors(t *testing.T) {
	loader := getLoader()

	t.Run("invalid_url",
		func(t *testing.T) {
			_, err := loader.Load(":foo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "missing protocol scheme") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("no_scheme",
		func(t *testing.T) {
			_, err := loader.Load("foo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "URL scheme not specified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("unknown_scheme",
		func(t *testing.T) {
			_, err := loader.Load("amqp://foo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "unknown URL scheme") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)
}

func TestPanic(t *testing.T) {
	t.Run("no drivers",
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

	t.Run("invalid type",
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
		},
	}

	return conf.NewLoader(driver)
}

func (d *testDriver) Name() string {
	return "test"
}

func (d *testDriver) Load(urlAddr *url.URL) (interface{}, error) {
	key := urlAddr.Host
	return d.sections[key], nil
}
