package conf_test

import (
	"reflect"
	"testing"

	"github.com/iph0/conf"
)

type updatesNotifier struct {
	updates chan<- string
}

func (n *updatesNotifier) Notify(provider string) {
	n.updates <- provider
}

func init() {
	conf.RegisterProvider("test",
		func() (conf.Provider, error) {
			provider := newProvider()
			return provider, nil
		},
	)
}

func TestLoad(t *testing.T) {
	updates := make(chan string)
	updNotifier := &updatesNotifier{updates}

	loader, err := conf.NewLoader(
		conf.LoaderConfig{
			Locators: []string{
				"test:dirs",
				"test:db",
				"test:unknown",
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
