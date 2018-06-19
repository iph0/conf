package conf_test

import (
	"errors"

	"github.com/iph0/conf"
)

type testProvider struct {
	layers          map[string]interface{}
	updatesNotifier conf.UpdatesNotifier
}

func newProvider() conf.Provider {
	return &testProvider{
		layers: map[string]interface{}{
			"dirs": map[string]interface{}{
				"myapp": map[string]interface{}{
					"mediaFormats": []string{"images", "audio", "video"},
					"pageTitles":   map[string]string{"@var": ".mediaFormats"},
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

					"servers": map[string]interface{}{
						"@include": []string{"test:servers"},
					},
				},
			},

			"db": map[string]interface{}{
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
}

func (p *testProvider) Watch(notifier conf.UpdatesNotifier) {
	p.updatesNotifier = notifier
}

func (p *testProvider) Load(locator string) (interface{}, error) {
	if locator == "invalid" {
		return nil, errors.New("loading failed")
	}

	config, ok := p.layers[locator]

	if !ok {
		return nil, nil
	}

	return config, nil
}

func (p *testProvider) Close() {}
