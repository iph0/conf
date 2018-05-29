package envconf_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/iph0/conf"
	"github.com/iph0/conf/envconf"
)

func init() {
	os.Setenv("MYAPP_ROOTDIR", "/myapp")
	os.Setenv("MYAPP_DBUSER", "stat_writer")
	os.Setenv("MYAPP_DBPASS", "stat_writer_pass")
}

func TestLoad(t *testing.T) {
	loader := conf.NewLoader(
		&envconf.EnvProvider{},
	)

	tConfig, err := loader.Load(
		"env:^MYAPP_.*",

		map[string]interface{}{
			"myapp": map[string]interface{}{
				"dirs": map[string]interface{}{
					"templatesDir": "${MYAPP_ROOTDIR}/templates",
					"sessionsDir":  "${MYAPP_ROOTDIR}/sessions",
				},

				"db": map[string]interface{}{
					"connectors": map[string]interface{}{
						"stat": map[string]interface{}{
							"host":     "localhost",
							"port":     1234,
							"dbname":   "stat",
							"username": map[string]interface{}{"@var": "MYAPP_DBUSER"},
							"password": map[string]interface{}{"@var": "MYAPP_DBPASS"},
						},
					},
				},
			},
		},
	)

	eConfig := map[string]interface{}{
		"myapp": map[string]interface{}{
			"dirs": map[string]interface{}{
				"templatesDir": "/myapp/templates",
				"sessionsDir":  "/myapp/sessions",
			},

			"db": map[string]interface{}{
				"connectors": map[string]interface{}{
					"stat": map[string]interface{}{
						"host":     "localhost",
						"port":     1234,
						"dbname":   "stat",
						"username": "stat_writer",
						"password": "stat_writer_pass",
					},
				},
			},
		},

		"MYAPP_ROOTDIR": "/myapp",
		"MYAPP_DBUSER":  "stat_writer",
		"MYAPP_DBPASS":  "stat_writer_pass",
	}

	if err != nil {
		t.Error("failed to load configuration:", err)
		return
	}

	if !reflect.DeepEqual(tConfig, eConfig) {
		t.Errorf("unexpected configuration returned: %#v", tConfig)
	}
}

func TestErrors(t *testing.T) {
	provider := &envconf.EnvProvider{}

	t.Run("empty_pattern",
		func(t *testing.T) {
			_, err := provider.Load("")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "empty pattern specified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("unknown_provider",
		func(t *testing.T) {
			_, err := provider.Load("redis:foo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "unknown pattern specified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_pattern",
		func(t *testing.T) {
			_, err := provider.Load("^MYAPP_[")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "error parsing regexp") == -1 {
				t.Error("other error happened:", err)
			}

			_, err = provider.Load("env:^MYAPP_[")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "error parsing regexp") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)
}
