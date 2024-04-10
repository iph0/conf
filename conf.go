package conf

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/iph0/merger"
	mapstruct "github.com/mitchellh/mapstructure"
)

const (
	errPref        = "conf"
	decoderTagName = "conf"
	refNameSep     = "."
)

// Processor loads configuration layers from different sources and merges them
// into the one configuration tree. In addition configuration processor can
// expand references on configuration parameters in string values and process
// $ref and $include directives in resulting configuration tree. Processing can
// be disabled if not needed.
type Processor struct {
	config ProcessorConfig
	stack  []string
	seen   map[reflect.Value]struct{}
	refs   map[string]reflect.Value
	root   reflect.Value
}

var (
	refKey          = reflect.ValueOf("$ref")
	includeKey      = reflect.ValueOf("$include")
	nameKey         = reflect.ValueOf("name")
	firstDefinedKey = reflect.ValueOf("firstDefined")
	defaultKey      = reflect.ValueOf("default")
)

// ProcessorConfig is a structure with configuration parameters for configuration
// processor.
type ProcessorConfig struct {
	// Loaders specifies configuration loaders. Map keys reperesents names of
	// configuration loaders, that further can be used in configuration locators.
	Loaders map[string]Loader

	// DisableProcessing disables expansion of references and processing of
	// directives.
	DisableProcessing bool
}

// Loader is an interface for configuration loaders.
type Loader interface {
	Load(string) ([]any, error)
}

// M type is a convenient alias for a map[string]any map.
type M = map[string]any

// A type is a convenient alias for a []any slice.
type A = []any

type walkFunc func(node *reflect.Value) error

// NewProcessor method creates new configuration processor instance.
func NewProcessor(config ProcessorConfig) *Processor {
	if config.Loaders == nil {
		config.Loaders = make(map[string]Loader)
	}

	return &Processor{
		config: config,
	}
}

// Decode method decodes raw configuration data into structure. Note that the
// conf tags defined in the struct type can indicate which fields the values are
// mapped to (see the example below). The decoder will make the following conversions:
//   - bools to string (true = "1", false = "0")
//   - numbers to string (base 10)
//   - bools to int/uint (true = 1, false = 0)
//   - strings to int/uint (base implied by prefix)
//   - int to bool (true if value != 0)
//   - string to bool (accepts: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false,
//     False. Anything else is an error)
//   - empty array = empty map and vice versa
//   - negative numbers to overflowed uint values (base 10)
//   - slice of maps to a merged map
//   - single values are converted to slices if required. Each element also can
//     be converted. For example: "4" can become []int{4} if the target type is
//     an int slice.
func Decode(configRaw, config any) error {
	decoder, err := mapstruct.NewDecoder(
		&mapstruct.DecoderConfig{
			WeaklyTypedInput: true,
			Result:           config,
			TagName:          decoderTagName,
		},
	)

	if err != nil {
		return err
	}

	err = decoder.Decode(configRaw)

	if err != nil {
		return err
	}

	return nil
}

// Load method loads configuration tree using configuration locators. The merge
// priority of loaded configuration layers depends on the order of configuration
// locators. Layers loaded by rightmost locator have highest priority.
func (p *Processor) Load(locators ...any) (M, error) {
	if len(locators) == 0 {
		panic(fmt.Errorf("%s: no configuration locators specified", errPref))
	}

	layers, err := p.load(locators)

	if err != nil {
		return nil, err
	}

	if !p.config.DisableProcessing {
		for i, layer := range layers {
			var err error
			layer, err = p.preprocess(layer)

			if err != nil {
				return nil, err
			}

			layers[i] = layer
		}
	}

	config := p.merge(layers)

	if config == nil {
		return nil, nil
	}

	if !p.config.DisableProcessing {
		var err error
		config, err = p.process(config)

		if err != nil {
			return nil, err
		}
	}

	if conf, ok := config.(M); ok {
		return conf, nil
	}

	return nil,
		fmt.Errorf("%s: loaded configuration must be a map of type conf.M, but got: %T",
			errPref, config)
}

func (p *Processor) load(locators []any) ([]any, error) {
	var allLayers []any

	for _, locator := range locators {
		switch loc := locator.(type) {
		case M:
			allLayers = append(allLayers, loc)
		case string:
			if loc == "" {
				return nil, fmt.Errorf("%s: empty configuration locator specified",
					errPref)
			}

			tokens := strings.SplitN(loc, ":", 2)

			if len(tokens) < 2 || tokens[0] == "" {
				return nil, fmt.Errorf("%s: missing loader name in configuration locator",
					errPref)
			}

			loaderName := tokens[0]
			locValue := tokens[1]

			if loader, ok := p.config.Loaders[loaderName]; ok {
				layers, err := loader.Load(locValue)

				if err != nil {
					return nil, err
				} else if len(layers) == 0 {
					continue
				}

				allLayers = append(allLayers, layers...)
			} else {
				return nil, fmt.Errorf("%s: unknown loader: %s", errPref, loaderName)
			}
		default:
			return nil,
				fmt.Errorf("%s: configuration locator must be a string or a map of type conf.M,"+
					" but got: %T", errPref, locator)
		}
	}

	return allLayers, nil
}

func (p *Processor) preprocess(layer any) (any, error) {
	p.stack = make([]string, 0, 10)
	p.seen = make(map[reflect.Value]struct{})

	defer func() {
		p.stack = nil
		p.seen = nil
	}()

	lyr := reflect.ValueOf(layer)
	lyr, err := p.applyInclude(lyr)

	if err != nil {
		return nil, err
	}

	p.root = lyr

	err = p.walk(p.root,
		func(nodePtr *reflect.Value) error {
			node, err := p.applyInclude(*nodePtr)

			if err != nil {
				return err
			}

			*nodePtr = node

			return nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("%s at %s", err, p.processContext())
	}

	return lyr.Interface(), nil
}

func (p *Processor) merge(layers []any) any {
	var config any

	for _, layer := range layers {
		config = merger.Merge(config, layer)
	}

	return config
}

func (p *Processor) process(config any) (any, error) {
	p.stack = make([]string, 0, 10)
	p.seen = make(map[reflect.Value]struct{})
	p.refs = make(map[string]reflect.Value)

	defer func() {
		p.stack = nil
		p.seen = nil
		p.refs = nil
		p.root = reflect.Value{}
	}()

	conf := reflect.ValueOf(config)
	conf, err := p.applyDirectives(conf)

	if err != nil {
		return nil, err
	}

	p.root = conf

	err = p.walk(p.root,
		func(nodePtr *reflect.Value) error {
			node, err := p.applyDirectives(*nodePtr)

			if err != nil {
				return err
			}

			*nodePtr = node

			return nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("%s at %s", err, p.processContext())
	}

	return conf.Interface(), nil
}

func (p *Processor) walk(node reflect.Value, f walkFunc) error {
	node = strip(node)

	switch node.Kind() {
	case reflect.Map:
		return p.walkMap(node, f)
	case reflect.Slice:
		return p.walkSlice(node, f)
	}

	return nil
}

func (p *Processor) walkMap(m reflect.Value, f walkFunc) error {
	if _, ok := p.seen[m]; ok {
		return nil
	}

	p.seen[m] = struct{}{}

	for _, key := range m.MapKeys() {
		keyStr := key.Interface().(string)
		p.pushToStack(keyStr)

		node := m.MapIndex(key)
		err := f(&node)

		if err != nil {
			return err
		}

		m.SetMapIndex(key, node)
		err = p.walk(node, f)

		if err != nil {
			return err
		}

		p.popFromStack()
	}

	return nil
}

func (p *Processor) walkSlice(s reflect.Value, f walkFunc) error {
	if _, ok := p.seen[s]; ok {
		return nil
	}

	p.seen[s] = struct{}{}

	sliceLen := s.Len()

	for i := 0; i < sliceLen; i++ {
		indexStr := strconv.Itoa(i)
		p.pushToStack(indexStr)

		node := s.Index(i)
		err := f(&node)

		if err != nil {
			return err
		}

		s.Index(i).Set(node)
		err = p.walk(node, f)

		if err != nil {
			return err
		}

		p.popFromStack()
	}

	return nil
}

func (p *Processor) applyInclude(node reflect.Value) (reflect.Value, error) {
	node = strip(node)

	switch node.Kind() {
	case reflect.Map:
		if locators := node.MapIndex(includeKey); locators.IsValid() {
			value, err := p.include(locators)

			if err != nil {
				return reflect.Value{}, err
			}

			return value, nil
		}
	}

	return node, nil
}

func (p *Processor) applyDirectives(node reflect.Value) (reflect.Value, error) {
	node = strip(node)

	switch node.Kind() {
	case reflect.String:
		str := node.Interface().(string)
		str, err := p.expandRefs(str)

		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(str), nil
	case reflect.Map:
		if ref := node.MapIndex(refKey); ref.IsValid() {
			value, err := p.resolveRef(ref)

			if err != nil {
				return reflect.Value{}, err
			}

			return value, nil
		}
	}

	return node, nil
}

func (p *Processor) expandRefs(str string) (string, error) {
	var res string
	runes := []rune(str)
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
				res += string(runes[i:j])

				for i, j = j, j+k+1; j < runesLen; j++ {
					if runes[j] == '}' {
						if esc {
							res += string(runes[i+1 : j+1])
						} else {
							name := string(runes[i+2 : j])

							if len(name) > 0 {
								node, err := p.fetchNode(name)

								if err != nil {
									return "", err
								}

								if node.IsValid() {
									res += fmt.Sprintf("%v", node.Interface())
								}
							} else {
								res += string(runes[i : j+1])
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

	res += string(runes[i:j])

	return res, nil
}

func (p *Processor) resolveRef(ref reflect.Value) (reflect.Value, error) {
	ref = strip(ref)

	switch ref.Kind() {
	case reflect.String:
		nameStr := ref.Interface().(string)
		node, err := p.fetchNode(nameStr)

		if err != nil {
			return reflect.Value{}, err
		}

		return node, nil
	case reflect.Map:
		if name := ref.MapIndex(nameKey); name.IsValid() {
			name = strip(name)
			nameKind := name.Kind()

			if nameKind != reflect.String {
				return reflect.Value{},
					fmt.Errorf("%s: reference name must be a string, but got: %s", errPref,
						nameKind)
			}

			nameStr := name.Interface().(string)
			node, err := p.fetchNode(nameStr)

			if err != nil {
				return reflect.Value{}, err
			}

			if node.IsValid() {
				return node, nil
			}
		} else if names := ref.MapIndex(firstDefinedKey); names.IsValid() {
			names = strip(names)
			namesKind := names.Kind()

			if namesKind != reflect.Slice {
				return reflect.Value{},
					fmt.Errorf("%s: '\"%s\" must be an array, but got: %s", errPref,
						firstDefinedKey, namesKind)
			}

			namesLen := names.Len()

			for i := 0; i < namesLen; i++ {
				name := names.Index(i)
				name = strip(name)
				nameKind := name.Kind()

				if nameKind != reflect.String {
					return reflect.Value{},
						fmt.Errorf("%s: reference name in \"%s\" must be a string, but got: %s",
							errPref, firstDefinedKey, nameKind)
				}

				nameStr := name.Interface().(string)
				node, err := p.fetchNode(nameStr)

				if err != nil {
					return reflect.Value{}, err
				}

				if node.IsValid() {
					return node, nil
				}
			}
		}

		node := ref.MapIndex(defaultKey)

		if node.IsValid() {
			return node, nil
		}
	}

	return reflect.Value{}, fmt.Errorf("%s: malformed directive: %s", errPref,
		refKey)
}

func (p *Processor) include(locators reflect.Value) (reflect.Value, error) {
	locators = strip(locators)
	locsKind := locators.Kind()

	if locsKind != reflect.Slice {
		return reflect.Value{},
			fmt.Errorf("%s: locators in %s directive must be specified as an array, but got: %s",
				errPref, includeKey, locsKind)
	}

	locsLen := locators.Len()
	locs := make([]any, locsLen)

	for i := 0; i < locsLen; i++ {
		loc := locators.Index(i)
		loc = strip(loc)
		locKind := loc.Kind()

		if locKind != reflect.String {
			return reflect.Value{},
				fmt.Errorf("%s: locator in %s directive must be a string, but got: %s",
					errPref, includeKey, locKind)
		}

		locs[i] = loc.Interface()
	}

	layers, err := p.load(locs)

	if err != nil {
		return reflect.Value{}, err
	}

	config := p.merge(layers)

	return reflect.ValueOf(config), nil
}

func (p *Processor) fetchNode(name string) (reflect.Value, error) {
	if node, ok := p.refs[name]; ok {
		return node, nil
	}

	node, err := p.findNode(name)

	if err != nil {
		return reflect.Value{}, err
	}

	p.refs[name] = node

	return node, nil
}

func (p *Processor) findNode(name string) (reflect.Value, error) {
	node := p.root
	tokens := strings.Split(name, refNameSep)
	tokensLen := len(tokens)

	for i := 0; i < tokensLen; i++ {
		node = strip(node)
		tokens[i] = strings.Trim(tokens[i], " ")

		switch node.Kind() {
		case reflect.Map:
			key := reflect.ValueOf(tokens[i])
			stackTemp := p.stack
			p.stack = tokens[:i+1]

			child := node.MapIndex(key)
			child, err := p.applyDirectives(child)

			if err != nil {
				return reflect.Value{}, err
			}

			node.SetMapIndex(key, child)
			node = child
			p.stack = stackTemp
		case reflect.Slice:
			j, err := strconv.Atoi(tokens[i])

			if err != nil {
				return reflect.Value{}, fmt.Errorf("%s: invalid array index: %s",
					errPref, tokens[i])
			} else if j < 0 || j >= node.Len() {
				return reflect.Value{}, fmt.Errorf("%s: array index out of range",
					errPref)
			}

			stackTemp := p.stack
			p.stack = tokens[:i+1]

			childNode := node.Index(j)
			childNode, err = p.applyDirectives(childNode)

			if err != nil {
				return reflect.Value{}, err
			}

			node.Index(j).Set(childNode)
			node = childNode
			p.stack = stackTemp
		default:
			return reflect.Value{}, nil
		}

		if !node.IsValid() {
			return reflect.Value{}, nil
		}
	}

	return node, nil
}

func strip(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Interface {
		return value.Elem()
	}

	return value
}

func (p *Processor) pushToStack(bc string) {
	p.stack = append(p.stack, bc)
}

func (p *Processor) popFromStack() {
	p.stack = p.stack[:len(p.stack)-1]
}

func (p *Processor) processContext() string {
	return strings.Join(p.stack, refNameSep)
}
