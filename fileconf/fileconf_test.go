package fileconf_test

import (
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
		fileconf.NewDriver(true),
	)

	tConfig, err := loader.Load(
		"file:dirs.yml",
		"file:db.json",

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
	driver := fileconf.NewDriver(true)

	t.Run("empty_pattern",
		func(t *testing.T) {
			_, err := driver.Load("")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "empty pattern specified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("no_driver",
		func(t *testing.T) {
			_, err := driver.Load("foo.yml")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "driver name not specified") == -1 {
				t.Error("other error happened:", err)
			}

			_, err = driver.Load(":foo.yml")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "driver name not specified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("unknown_driver",
		func(t *testing.T) {
			_, err := driver.Load("redis:foo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "unknown driver name") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_pattern",
		func(t *testing.T) {
			_, err := driver.Load("file:dirs.y[*ml")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "syntax error in pattern") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("not_found",
		func(t *testing.T) {
			_, err := driver.Load("file:unknown.yml")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "nothing found") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("no_extension",
		func(t *testing.T) {
			_, err := driver.Load("file:foo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "file extension not specified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("unknown_extension",
		func(t *testing.T) {
			_, err := driver.Load("file:bar.xml")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "unknown file extension") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)
}
