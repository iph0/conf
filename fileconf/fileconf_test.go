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
	os.Setenv("GOCONF_PATH", "./etc")
}

func TestLoad(t *testing.T) {
	loader, err := NewLoader()

	if err != nil {
		t.Error(err)
		return
	}

	tConfig, err := loader.Load(
		map[string]interface{}{
			"paramA": "default:valA",
			"paramZ": "default:valZ",
		},

		"file:foo.yml",
		"file:bar.json",
	)

	if err != nil {
		t.Error(err)
		return
	}

	eConfig := map[string]interface{}{
		"paramA": "foo:valA",
		"paramB": "bar:valB",
		"paramC": "bar:valC",

		"paramD": map[string]interface{}{
			"paramDA": "foo:valDA",
			"paramDB": "bar:valDB",
			"paramDC": "bar:valDC",
			"paramDE": "foo:bar:valDC",

			"paramDF": []interface{}{
				"foo:valDFA",
				"foo:valDFB",
				"foo:foo:valDA",
			},
		},

		"paramE": []interface{}{
			"bar:valEA",
			"bar:valEB",
		},

		"paramF": "foo:bar:valB",
		"paramG": "bar:foo:valDA",
		"paramH": "foo:bar:valEA",
		"paramI": "bar:foo:bar:valEA",
		"paramJ": "foo:bar:foo:bar:valEA",
		"paramK": "bar:foo:valDFB:foo:bar:valDC",
		"paramL": "foo:${paramD.paramDE}:${}:${paramD.paramDA}",

		"paramM": map[string]interface{}{
			"paramDA": "foo:valDA",
			"paramDB": "bar:valDB",
			"paramDC": "bar:valDC",
			"paramDE": "foo:bar:valDC",

			"paramDF": []interface{}{
				"foo:valDFA",
				"foo:valDFB",
				"foo:foo:valDA",
			},
		},

		"paramN": map[string]interface{}{
			"paramNA": "foo:valNA",
			"paramNB": "foo:valNB",

			"paramNC": map[string]interface{}{
				"paramNCA": "foo:valNCA",
				"paramNCB": "bar:valNCB",
				"paramNCC": "bar:valNCC",
				"paramNCD": "bar:foo:valNCA",
				"paramNCE": "foo:valNB",
			},
		},

		"paramO": map[string]interface{}{
			"paramOA": "moo:valOA",
			"paramOB": "jar:valOB",
			"paramOC": "jar:valOC",

			"paramOD": map[string]interface{}{
				"paramODA": "moo:valODA",
				"paramODB": "jar:valODB",
				"paramODC": "jar:valODC",
				"paramODD": "jar:bar:valNCB",
			},

			"paramOE": []interface{}{
				"zoo:valA",
				"zoo:valB",
			},
		},

		"paramP": map[string]interface{}{
			"paramODA": "moo:valODA",
			"paramODB": "jar:valODB",
			"paramODC": "jar:valODC",
			"paramODD": "jar:bar:valNCB",
		},

		"paramZ": "default:valZ",
	}

	if !reflect.DeepEqual(tConfig, eConfig) {
		t.Errorf("unexpected configuration returned: %#v", tConfig)
	}
}

func TestErrors(t *testing.T) {
	loader, err := NewLoader()

	if err != nil {
		t.Error(err)
		return
	}

	t.Run("file_extension_not_specified",
		func(t *testing.T) {
			_, err := loader.Load("file:coo")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "file extension not specified") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("unknown_file_extension",
		func(t *testing.T) {
			_, err := loader.Load("file:mar.html")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "unknown file extension") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("invalid_pattern",
		func(t *testing.T) {
			_, err := loader.Load("file:f[oo.yml")

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "syntax error in pattern") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)

	t.Run("GOCONF_PATH_not_set",
		func(t *testing.T) {
			os.Setenv("GOCONF_PATH", "")
			_, err := NewLoader()

			if err == nil {
				t.Error("no error happened")
			} else if strings.Index(err.Error(), "GOCONF_PATH not set") == -1 {
				t.Error("other error happened:", err)
			}
		},
	)
}

func NewLoader() (*conf.Loader, error) {
	fileProv, err := fileconf.NewProvider()

	if err != nil {
		return nil, err
	}

	loader := conf.NewLoader(
		map[string]conf.Provider{
			"file": fileProv,
		},
	)

	return loader, nil
}
