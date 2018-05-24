package fileconf_test

import (
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/iph0/conf"
	"github.com/iph0/conf/fileconf"
)

func init() {
	os.Setenv("GOCONF_PATH", "fileconf_test/etc")
}

func TestLoad(t *testing.T) {
	loader := conf.NewLoader(
		fileconf.NewLoaderDriver(true),
	)

	tConfig, err := loader.Load(
		"file:///dirs.yml",
		"file:///db.json",

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
		t.Error(err)
		return
	}

	eConfig := map[string]interface{}{
		"myapp": map[string]interface{}{
			"mediaFormats": []interface{}{"images", "audio", "video"},
			"metadata":     "foo:${moo.jar}:bar",

			"dirs": map[string]interface{}{
				"rootDir":      "/myapp",
				"templatesDir": "/myapp/templates", "sessionsDir": "/myapp/sessions",
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
						"port":     float64(1234),
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
	driver := fileconf.NewLoaderDriver(true)
	loader := conf.NewLoader(driver)

	t.Run("no_scheme",
		func(t *testing.T) {
			uriAddr, _ := url.Parse("dirs.yml")
			_, err := driver.Load(uriAddr)

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "URI scheme not specified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("unknown_scheme",
		func(t *testing.T) {
			driver := fileconf.NewLoaderDriver(true)
			uriAddr, _ := url.Parse("amqp://foo")
			_, err := driver.Load(uriAddr)

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "unknown URI scheme") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("no_path",
		func(t *testing.T) {
			_, err := loader.Load("file://dirs.yml")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "URI path not specified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("not_found",
		func(t *testing.T) {
			_, err := loader.Load("file:///unknown.yml")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "configuration data not found") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("no_extension",
		func(t *testing.T) {
			_, err := loader.Load("file:///foo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "file extension not specified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("unknown_extension",
		func(t *testing.T) {
			_, err := loader.Load("file:///bar.xml")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "unknown file extension") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_pattern",
		func(t *testing.T) {
			_, err := loader.Load("file:///dirs.y[*ml")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "syntax error in pattern") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_escape",
		func(t *testing.T) {
			_, err := loader.Load("file://%3F")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "invalid URL escape") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)
}
