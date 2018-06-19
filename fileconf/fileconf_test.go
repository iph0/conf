package fileconf_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/iph0/conf"
	"github.com/iph0/conf/fileconf"
)

type updatesNotifier struct {
	updates chan<- string
}

func (n *updatesNotifier) Notify(provider string) {
	n.updates <- provider
}

func init() {
	conf.RegisterProvider("file",
		func() (conf.Provider, error) {
			provider := fileconf.NewProvider()
			return provider, nil
		},
	)

	os.Setenv("GOCONF_PATH", "fileconf_test/etc")
}

func TestLoad(t *testing.T) {
	updates := make(chan string)
	updNotifier := &updatesNotifier{updates}

	loader, err := conf.NewLoader(
		conf.LoaderConfig{
			Locators: []string{
				"file:dirs.yml",
				"file:db.json",
			},

			Watch: updNotifier,
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
		"myapp": map[string]interface{}{
			"mediaFormats": []interface{}{"images", "audio", "video"},
			"pageTitles":   []interface{}{"images", "audio", "video"},
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
						"host":     "stat.mydb.com",
						"port":     float64(4321),
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
