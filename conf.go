package conf

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/iph0/merger"
)

const (
	errPref    = "conf"
	varNameSep = "."
)

var (
	provConstrs = make(map[string]func() (Provider, error))

	varKey     = reflect.ValueOf("@var")
	includeKey = reflect.ValueOf("@include")
	emptyStr   = reflect.ValueOf("")
	zero       = reflect.ValueOf(nil)
)

// Loader TODO
type Loader struct {
	providers   map[string]Provider
	locators    []*locator
	root        reflect.Value
	breadcrumbs []string
	vars        map[string]reflect.Value
	seen        map[reflect.Value]bool
}

// LoaderConfig TODO
type LoaderConfig struct {
	Locators []string
	Watch    UpdatesNotifier
}

// UpdatesNotifier is an interface for update notifiers.
type UpdatesNotifier interface {
	Notify(provider string)
}

// Provider is an interface for configuration providers.
type Provider interface {
	Watch(UpdatesNotifier)
	Load(string) (interface{}, error)
	Close()
}

// RegisterProvider method registers constructor for configuration provider.
func RegisterProvider(name string, constr func() (Provider, error)) {
	provConstrs[name] = constr
}

// NewLoader TODO
func NewLoader(config LoaderConfig) (*Loader, error) {
	if len(config.Locators) == 0 {
		return nil, fmt.Errorf("%s: no configuration locators specified", errPref)
	}

	provs := make(map[string]Provider)
	locs := make([]*locator, 0, len(config.Locators))

	for _, rawLoc := range config.Locators {
		loc, err := parseLocator(rawLoc)

		if err != nil {
			return nil, err
		}

		provConstr, ok := provConstrs[loc.Provider]

		if !ok {
			return nil,
				fmt.Errorf("%s: provider not found for configuration locator %s",
					errPref, loc)
		}

		if _, ok := provs[loc.Provider]; !ok {
			prov, err := provConstr()

			if err != nil {
				return nil, err
			}

			if config.Watch != nil {
				prov.Watch(config.Watch)
			}

			provs[loc.Provider] = prov
		}

		locs = append(locs, loc)
	}

	return &Loader{
		providers: provs,
		locators:  locs,
	}, nil
}

// Load TODO
func (l *Loader) Load() (map[string]interface{}, error) {
	iConfig, err := l.load(l.locators)

	if err != nil {
		return nil, err
	}

	if iConfig == nil {
		return nil, nil
	}

	iConfig, err = l.process(iConfig)

	if err != nil {
		return nil, err
	}

	switch config := iConfig.(type) {
	case map[string]interface{}:
		return config, nil
	default:
		return nil, fmt.Errorf("%s: loaded configuration has invalid type %T",
			errPref, config)
	}
}

// Close method performs correct closure of the configuration keeper.
func (l *Loader) Close() {
	for _, provider := range l.providers {
		provider.Close()
	}
}

func (l *Loader) load(locs []*locator) (interface{}, error) {
	var layer interface{}

	for _, loc := range locs {
		provider, ok := l.providers[loc.Provider]

		if !ok {
			return nil,
				fmt.Errorf("%s: provider not found for configuration lacator %s",
					errPref, loc)
		}

		subLayer, err := provider.Load(loc.bareLocator)

		if err != nil {
			return nil, err
		}

		if subLayer == nil {
			continue
		}

		layer = merger.Merge(layer, subLayer)
	}

	return layer, nil
}

func (l *Loader) process(config interface{}) (interface{}, error) {
	root := reflect.ValueOf(config)
	l.root = root
	l.breadcrumbs = make([]string, 0, 10)
	l.vars = make(map[string]reflect.Value)
	l.seen = make(map[reflect.Value]bool)

	root, err := l.processNode(root)

	if err != nil {
		return nil, fmt.Errorf("%s at %s", err, l.errContext())
	}

	l.root = root
	err = l.walk(root)

	if err != nil {
		return nil, fmt.Errorf("%s at %s", err, l.errContext())
	}

	config = root.Interface()

	return config, nil
}

func (l *Loader) walk(node reflect.Value) error {
	node = stripValue(node)
	nodeKind := node.Kind()

	if nodeKind == reflect.Map ||
		nodeKind == reflect.Slice {

		if _, ok := l.seen[node]; ok {
			return nil
		}

		l.seen[node] = true
		var err error

		if nodeKind == reflect.Map {
			err = l.walkMap(node)
		} else { // Slice
			err = l.walkSlice(node)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Loader) walkMap(m reflect.Value) error {
	for _, key := range m.MapKeys() {
		iKey := key.Interface()
		l.pushCrumb(iKey.(string))

		value := m.MapIndex(key)
		value, err := l.processNode(value)

		if err != nil {
			return err
		}

		m.SetMapIndex(key, value)
		err = l.walk(value)

		if err != nil {
			return err
		}

		l.popCrumb()
	}

	return nil
}

func (l *Loader) walkSlice(s reflect.Value) error {
	sliceLen := s.Len()

	for i := 0; i < sliceLen; i++ {
		indexStr := strconv.Itoa(i)
		l.pushCrumb(indexStr)

		value := s.Index(i)
		value, err := l.processNode(value)

		if err != nil {
			return err
		}

		s.Index(i).Set(value)
		err = l.walk(value)

		if err != nil {
			return err
		}

		l.popCrumb()
	}

	return nil
}

func (l *Loader) processNode(node reflect.Value) (reflect.Value, error) {
	node = stripValue(node)
	nodeKind := node.Kind()
	var err error

	if nodeKind == reflect.String {
		node, err = l.processString(node)
	} else if nodeKind == reflect.Map {
		if name := node.MapIndex(varKey); name.IsValid() {
			node, err = l.processVar(name)
		} else if locs := node.MapIndex(includeKey); locs.IsValid() {
			node, err = l.processInclude(locs)
		}
	}

	if err != nil {
		return zero, err
	}

	return node, nil
}

func (l *Loader) processString(orig reflect.Value) (reflect.Value, error) {
	var resultStr string
	iOrig := orig.Interface()
	runes := []rune(iOrig.(string))
	runesLen := len(runes)
	i, j := 0, 0

	for j < runesLen {
		if runes[j] == '$' && j+1 < runesLen {
			var esc bool
			k := 1

			if runes[j+1] == '$' {
				esc = true
				k++
			}

			if runes[j+k] == '{' {
				resultStr += string(runes[i:j])

				for i, j = j, j+k+1; j < runesLen; j++ {
					if runes[j] == '}' {
						if esc {
							resultStr += string(runes[i+1 : j+1])
						} else {
							name := string(runes[i+2 : j])
							value, err := l.resolveVar(name)

							if err != nil {
								return zero, err
							}

							resultStr += fmt.Sprintf("%v", value.Interface())
						}

						i, j = j+1, j+1

						break
					}
				}

				continue
			}
		}

		j++
	}

	resultStr += string(runes[i:j])
	result := reflect.ValueOf(resultStr)

	return result, nil
}

func (l *Loader) processVar(name reflect.Value) (reflect.Value, error) {
	name = stripValue(name)
	nameKind := name.Kind()

	if nameKind != reflect.String {
		return zero, fmt.Errorf("%s: invalid @var directive", errPref)
	}

	iName := name.Interface()
	value, err := l.resolveVar(iName.(string))

	if err != nil {
		return zero, err
	}

	return value, nil
}

func (l *Loader) processInclude(rawLocs reflect.Value) (reflect.Value, error) {
	rawLocs = stripValue(rawLocs)
	locsKind := rawLocs.Kind()

	if locsKind != reflect.Slice {
		return zero, fmt.Errorf("%s: invalid @include directive", errPref)
	}

	locsLen := rawLocs.Len()
	locs := make([]*locator, 0, locsLen)

	for i := 0; i < locsLen; i++ {
		rawLoc := rawLocs.Index(i)
		rawLoc = stripValue(rawLoc)
		locKind := rawLoc.Kind()

		if locKind != reflect.String {
			return zero,
				fmt.Errorf("%s: configuration locator has invalid type %T", errPref,
					rawLoc.Interface())
		}

		iLoc := rawLoc.Interface()
		loc, err := parseLocator(iLoc.(string))

		if err != nil {
			return zero, err
		}

		locs = append(locs, loc)
	}

	layer, err := l.load(locs)

	if err != nil {
		return zero, err
	}

	return reflect.ValueOf(layer), nil
}

func (l *Loader) resolveVar(name string) (reflect.Value, error) {
	if name == "" {
		return l.root, nil
	}

	nameTokens := strings.Split(name, varNameSep)

	if nameTokens[0] == "" {
		nameTokens = l.toAbsVarName(nameTokens)
		name = strings.Join(nameTokens, varNameSep)
	}

	value, ok := l.vars[name]

	if ok {
		return value, nil
	}

	value, err := l.fetchVarValue(nameTokens)

	if err != nil {
		return zero, err
	}

	l.vars[name] = value

	return value, nil
}

func (l *Loader) toAbsVarName(nameTokens []string) []string {
	nameTokensLen := len(nameTokens)
	crumbsLen := len(l.breadcrumbs)
	i := 0

	for ; i < nameTokensLen; i++ {
		if nameTokens[i] != "" {
			break
		}
	}

	if i == nameTokensLen {
		i--

		if i >= crumbsLen {
			return l.breadcrumbs[:0]
		}

		return l.breadcrumbs[:crumbsLen-i]
	}

	if i >= crumbsLen {
		return nameTokens[i:]
	}

	return append(append([]string{}, l.breadcrumbs[:crumbsLen-i]...),
		nameTokens[i:]...)
}

func (l *Loader) fetchVarValue(nameTokens []string) (reflect.Value, error) {
	var node reflect.Value
	value := l.root
	nameTokensLen := len(nameTokens)
	i := 0

	for ; i < nameTokensLen; i++ {
		nameTokens[i] = strings.Trim(nameTokens[i], " ")
		value = stripValue(value)
		valueKind := value.Kind()

		if valueKind == reflect.Map {
			node = value
			key := reflect.ValueOf(nameTokens[i])

			crumbs := l.breadcrumbs
			l.breadcrumbs = nameTokens[:i+1]

			var err error
			value = node.MapIndex(key)
			value, err = l.processNode(value)

			l.breadcrumbs = crumbs

			if err != nil {
				return zero, err
			}

			node.SetMapIndex(key, value)
		} else if valueKind == reflect.Slice {
			node = value
			j, err := strconv.Atoi(nameTokens[i])

			if err != nil {
				return zero, fmt.Errorf("%s: invalid slice index", errPref)
			} else if j < 0 || j >= node.Len() {
				return zero, fmt.Errorf("%s: slice index out of range", errPref)
			}

			crumbs := l.breadcrumbs
			l.breadcrumbs = nameTokens[:i+1]

			value = node.Index(j)
			value, err = l.processNode(value)

			l.breadcrumbs = crumbs

			if err != nil {
				return zero, err
			}

			node.Index(j).Set(value)
		} else {
			return emptyStr, nil
		}

		if !value.IsValid() {
			return emptyStr, nil
		}
	}

	return value, nil
}

func (l *Loader) pushCrumb(bc string) {
	l.breadcrumbs = append(l.breadcrumbs, bc)
}

func (l *Loader) popCrumb() {
	l.breadcrumbs = l.breadcrumbs[:len(l.breadcrumbs)-1]
}

func stripValue(value reflect.Value) reflect.Value {
	valueKind := value.Kind()

	if valueKind == reflect.Interface {
		return value.Elem()
	}

	return value
}

func (l *Loader) errContext() string {
	return strings.Join(l.breadcrumbs, varNameSep)
}
