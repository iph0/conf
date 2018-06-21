package conf

import "fmt"

type mapProvider struct {
	layers map[string]interface{}
}

// NewMapProvider TODO
func NewMapProvider(layers map[string]interface{}) Provider {
	return &mapProvider{layers}
}

func (p *mapProvider) Watch(n UpdatesNotifier) {}

func (p *mapProvider) Load(loc *Locator) (interface{}, error) {
	key := loc.BareLocator
	layer, ok := p.layers[key]

	if !ok {
		return nil, fmt.Errorf("%s: configuration layer not found by locator %s",
			errPref, loc)
	}

	return layer, nil
}

func (p *mapProvider) Close() {}
