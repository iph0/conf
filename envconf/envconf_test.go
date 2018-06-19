package envconf_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/iph0/conf"
	"github.com/iph0/conf/envconf"
)

type updatesNotifier struct {
	updates chan<- string
}

func (n *updatesNotifier) Notify(provider string) {
	n.updates <- provider
}

func init() {
	os.Setenv("MYAPP_ROOTDIR", "/myapp")
	os.Setenv("MYAPP_DBUSER", "stat_writer")
	os.Setenv("MYAPP_DBPASS", "stat_writer_pass")

	conf.RegisterProvider("env",
		func() (conf.Provider, error) {
			provider := envconf.NewProvider()
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
				"env:^MYAPP_.*",
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
		"MYAPP_ROOTDIR": "/myapp",
		"MYAPP_DBUSER":  "stat_writer",
		"MYAPP_DBPASS":  "stat_writer_pass",
	}

	if !reflect.DeepEqual(tConfig, eConfig) {
		t.Errorf("unexpected configuration returned: %#v", tConfig)
	}
}
