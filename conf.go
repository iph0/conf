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
	varKey     = reflect.ValueOf("_var")
	includeKey = reflect.ValueOf("_include")
	emptyStr   = reflect.ValueOf("")
	zero       = reflect.ValueOf(nil)
)

// Loader loads configuration layers from different sources and merges them in
// one configuration tree. Also Loader performs expansion of variables and
// processing of special directives. Processing can be disabled if not needed.
type Loader struct {
	config      LoaderConfig
	root        reflect.Value
	breadcrumbs []string
	vars        map[string]reflect.Value
	seen        map[reflect.Value]bool
}

// LoaderConfig is a structure with configuration parameters for Loader.
type LoaderConfig struct {
	// Providers specifies a map of configuration providers. Keys in the map
	// reperesents names of configuration providers, that further must be used in
	// configuration locators.
	Providers map[string]Provider

	// DisableProcessing disables expansion of variables and processing of
	// directives.
	DisableProcessing bool
}

// Provider is an interface for configuration providers.
type Provider interface {
	Load(*Locator) (interface{}, error)
}

// NewLoader method creates new Loader instance.
func NewLoader(config LoaderConfig) *Loader {
	if config.Providers == nil {
		config.Providers = make(map[string]Provider)
	}

	return &Loader{
		config: config,
	}
}

// Load method loads configuration using configuration locators. The merge
// priority of loaded configuration layers depends on the order of configuration
// locators. Layers loaded by rightmost locator have highest priority.
func (l *Loader) Load(locators ...interface{}) (map[string]interface{}, error) {
	if len(locators) == 0 {
		panic(fmt.Errorf("%s: no configuration locators specified", errPref))
	}

	iConfig, err := l.load(locators)

	if err != nil {
		return nil, err
	}

	if iConfig == nil {
		return nil, nil
	}

	if !l.config.DisableProcessing {
		iConfig, err = l.process(iConfig)
	}

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

func (l *Loader) load(locators []interface{}) (interface{}, error) {
	var layer interface{}

	for _, iRawLoc := range locators {
		switch rawLoc := iRawLoc.(type) {
		case map[string]interface{}:
			layer = merger.Merge(layer, rawLoc)
		case string:
			loc, err := ParseLocator(rawLoc)

			if err != nil {
				return nil, err
			}

			prov, ok := l.config.Providers[loc.Provider]

			if !ok {
				return nil,
					fmt.Errorf("%s: provider not found for configuration locator %s",
						errPref, loc)
			}

			subLayer, err := prov.Load(loc)

			if err != nil {
				return nil, err
			}

			if subLayer == nil {
				continue
			}

			layer = merger.Merge(layer, subLayer)
		default:
			return nil, fmt.Errorf("%s: configuration locator has invalid type %T",
				errPref, rawLoc)
		}
	}

	return layer, nil
}

func (l *Loader) process(config interface{}) (interface{}, error) {
	root := reflect.ValueOf(config)
	l.root = root
	l.breadcrumbs = make([]string, 0, 10)
	l.vars = make(map[string]reflect.Value)
	l.seen = make(map[reflect.Value]bool)

	defer func() {
		l.root = zero
		l.breadcrumbs = nil
		l.vars = nil
		l.seen = nil
	}()

	root, err := l.processNode(root)

	if err != nil {
		return nil, err
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
	node = revealValue(node)
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
	node = revealValue(node)
	nodeKind := node.Kind()
	var err error

	if nodeKind == reflect.String {
		node, err = l.expandVars(node)
	} else if nodeKind == reflect.Map {
		if name := node.MapIndex(varKey); name.IsValid() {
			node, err = l.getVarValue(name)
		} else if locators := node.MapIndex(includeKey); locators.IsValid() {
			node, err = l.include(locators)
		}
	}

	if err != nil {
		return zero, err
	}

	return node, nil
}

func (l *Loader) expandVars(orig reflect.Value) (reflect.Value, error) {
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

							if len(name) > 0 {
								value, err := l.resolveVar(name)

								if err != nil {
									return zero, err
								}

								resultStr += fmt.Sprintf("%v", value.Interface())
							} else {
								resultStr += string(runes[i : j+1])
							}
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

func (l *Loader) getVarValue(name reflect.Value) (reflect.Value, error) {
	name = revealValue(name)
	nameKind := name.Kind()

	if nameKind != reflect.String {
		return zero, fmt.Errorf("%s: invalid _var directive", errPref)
	}

	iName := name.Interface()
	value, err := l.resolveVar(iName.(string))

	if err != nil {
		return zero, err
	}

	return value, nil
}

func (l *Loader) include(locators reflect.Value) (reflect.Value, error) {
	locators = revealValue(locators)
	locsKind := locators.Kind()

	if locsKind != reflect.Slice {
		return zero, fmt.Errorf("%s: invalid _include directive", errPref)
	}

	iLocators := locators.Interface()
	locsSlice := iLocators.([]interface{})
	layer, err := l.load(locsSlice)

	if err != nil {
		return zero, err
	}

	return reflect.ValueOf(layer), nil
}

func (l *Loader) resolveVar(name string) (reflect.Value, error) {
	if name[0] == '.' {
		nameLen := len(name)
		crumbsLen := len(l.breadcrumbs)
		i := 0

		for ; i < nameLen; i++ {
			if name[i] != '.' {
				break
			}
		}

		if i >= crumbsLen {
			name = name[i:]
		} else {
			baseName := strings.Join(l.breadcrumbs[:crumbsLen-i], varNameSep)

			if i == nameLen {
				name = baseName
			} else {
				name = baseName + varNameSep + name[i:]
			}
		}

		if name == "" {
			return l.root, nil
		}
	}

	value, ok := l.vars[name]

	if ok {
		return value, nil
	}

	value, err := l.findVarValue(name)

	if err != nil {
		return zero, err
	}

	l.vars[name] = value

	return value, nil
}

func (l *Loader) findVarValue(name string) (reflect.Value, error) {
	var node reflect.Value
	value := l.root
	tokens := strings.Split(name, varNameSep)
	tokensLen := len(tokens)
	i := 0

	for ; i < tokensLen; i++ {
		tokens[i] = strings.Trim(tokens[i], " ")
		value = revealValue(value)
		valueKind := value.Kind()

		if valueKind == reflect.Map {
			node = value
			key := reflect.ValueOf(tokens[i])

			crumbs := l.breadcrumbs
			l.breadcrumbs = tokens[:i+1]

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
			j, err := strconv.Atoi(tokens[i])

			if err != nil {
				return zero, fmt.Errorf("%s: invalid slice index", errPref)
			} else if j < 0 || j >= node.Len() {
				return zero, fmt.Errorf("%s: slice index out of range", errPref)
			}

			crumbs := l.breadcrumbs
			l.breadcrumbs = tokens[:i+1]

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

func revealValue(value reflect.Value) reflect.Value {
	valueKind := value.Kind()

	if valueKind == reflect.Interface {
		return value.Elem()
	}

	return value
}

func (l *Loader) errContext() string {
	return strings.Join(l.breadcrumbs, varNameSep)
}
