package conf_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/iph0/conf"
)

type driver struct {
	sections map[string]map[string]interface{}
}

func TestLoad(t *testing.T) {
	loader := getLoader()

	confTest, err := loader.Load("dirs", "db",
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
		t.Error("loading of configuration failed")
	}

	confExp := map[string]interface{}{
		"myapp": map[string]interface{}{
			"root_dir":      "/myapp",
			"templates_dir": "/myapp/templates", "sessions_dir": "/myapp/sessions",
			"media_dirs": []interface{}{
				"/myapp/media/images",
				"/myapp/media/audio",
				"/myapp/media/video"},

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
						"port":     1234,
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

	if !reflect.DeepEqual(confTest, confExp) {
		t.Errorf("unexpected configuration returned: %#v", confTest)
	}
}

func TestNotFound(t *testing.T) {
	loader := getLoader()

	_, err := loader.Load("unknown")

	if err == nil {
		t.Error("unexpected behavior detected")
	}
}

func TestPanic(t *testing.T) {
	t.Run("no drivers",
		func(t *testing.T) {
			defer func() {
				err := recover()

				if err == nil {
					t.Error("no error happened")
				} else if strings.Index(err.(string), "no drivers specified") == -1 {
					t.Error("other error happened")
				}
			}()

			conf.NewLoader()
		},
	)

	t.Run("invalid type",
		func(t *testing.T) {
			defer func() {
				err := recover()

				if err == nil {
					t.Error("no error happened")
				} else if strings.Index(err.(string), "is invalid type") == -1 {
					t.Error("other error happened")
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
	var driverInst = &driver{
		map[string]map[string]interface{}{
			"dirs": {
				"myapp": map[string]interface{}{
					"root_dir":      "/myapp",
					"templates_dir": "/myapp/templates",
					"sessions_dir":  "/myapp/sessions",
					"media_dirs": []interface{}{
						"/myapp/media/images",
						"/myapp/media/audio",
						"/myapp/media/video",
					},
				},
			},

			"db": {
				"myapp": map[string]interface{}{
					"db": map[string]interface{}{
						"connectors": map[string]interface{}{
							"stat_master": map[string]interface{}{
								"host":     "stat-master.mydb.com",
								"port":     1234,
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
								"port":     1234,
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
			},
		},
	}

	return conf.NewLoader(driverInst)
}

func (d *driver) Load(sec string) (map[string]interface{}, error) {
	return d.sections[sec], nil
}
