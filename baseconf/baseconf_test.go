package baseconf_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/iph0/conf"
	"github.com/iph0/conf/baseconf"
)

func init() {
	os.Setenv("GOCONF_PATH", "baseconf_test/etc")
}

func TestLoad(t *testing.T) {
	loader := conf.NewLoader(
		baseconf.NewDriver(),
	)

	tConf, err := loader.Load("dirs", "db",
		map[string]interface{}{
			"myapp": map[string]interface{}{
				"db": map[string]interface{}{
					"connectors": map[string]interface{}{
						"stat_master": map[string]interface{}{
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
	}

	eConf := map[string]interface{}{
		"myapp": map[string]interface{}{
			"root_dir":      "/myapp",
			"templates_dir": "/myapp/templates", "sessions_dir": "/myapp/sessions",
			"media_dirs": []interface{}{
				"/myapp/media/images",
				"/myapp/media/audio",
				"/myapp/media/video",
			},

			"db": map[string]interface{}{
				"connectors": map[string]interface{}{
					"stat_master": map[string]interface{}{
						"host":     "localhost",
						"port":     4321,
						"dbname":   "stat",
						"username": "stat_writer",
						"password": "stat_writer_pass",
						"options": map[string]interface{}{
							"PrintWarn":  false,
							"PrintError": false,
							"RaiseError": true,
						},
					},

					"stat_slave": map[string]interface{}{
						"host":     "stat-slave.mydb.com",
						"port":     float64(1234),
						"dbname":   "stat",
						"username": "stat_reader",
						"password": "stat_reader_pass",

						"options": map[string]interface{}{
							"PrintWarn":  false,
							"PrintError": false,
							"RaiseError": true,
						},
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(tConf, eConf) {
		t.Errorf("unexpected configuration returned: %#v", tConf)
	}
}

func TestNotFound(t *testing.T) {
	loader := conf.NewLoader(
		baseconf.NewDriver(),
	)

	tConf, err := loader.Load("unknown")

	if tConf != nil && err == nil {
		t.Error("unexpected behavior detected")
	}
}
