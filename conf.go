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
	keyStackCap    = 10
)

// Processor loads configuration layers from different sources and merges them
// into the one configuration tree. In addition configuration processor can
// expand references on configuration parameters in string values and process
// $ref and $include directives in resulting configuration tree. Processing can
// be disabled if not needed.
type Processor struct {
	config    ProcessorConfig
	keyStack  *keyStack
	seenNodes map[uintptr]struct{}
	refs      map[string]reflect.Value
	root      reflect.Value
}

var (
	refKey      = reflect.ValueOf("$ref")
	includeKey  = reflect.ValueOf("$include")
	underlayKey = reflect.ValueOf("$underlay")
	overlayKey  = reflect.ValueOf("$overlay")

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

type keyStack struct {
	s []string
}

type processFunc func(node reflect.Value) (reflect.Value, error)

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
			layer, err := p.processLayer(layer)

			if err != nil {
				return nil, err
			}

			layers[i] = layer
		}
	}

	var config any

	for _, layer := range layers {
		config = merger.Merge(config, layer)
	}

	if config == nil {
		return nil, nil
	}

	if !p.config.DisableProcessing {
		var err error
		config, err = p.processConfig(config)

		if err != nil {
			return nil, err
		}
	}

	if conf, ok := config.(M); ok {
		return conf, nil
	}

	return nil,
		fmt.Errorf("%s: loaded configuration must be of type \"map[string]any\", "+
			"but got \"%T\"", errPref, config)
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
				fmt.Errorf("%s: configuration locator must be of type \"string\" or "+
					"\"map[string]any\", but got \"%T\"", errPref, locator)
		}
	}

	return allLayers, nil
}

func (p *Processor) processLayer(layer any) (any, error) {
	p.beforeProcess()
	defer p.afterProcess()

	lyr := reflect.ValueOf(layer)
	lyr, err := p.processIncludes(lyr)

	if err != nil {
		return nil, err
	}

	return lyr.Interface(), nil
}

func (p *Processor) processConfig(config any) (any, error) {
	p.beforeProcess()
	defer p.afterProcess()

	conf := reflect.ValueOf(config)
	p.root = conf
	conf, err := p.processDirectives(conf)

	if err != nil {
		return nil, err
	}

	return conf.Interface(), nil
}

func (p *Processor) processIncludes(node reflect.Value) (reflect.Value, error) {
	return p.processNode(node,
		func(node reflect.Value) (reflect.Value, error) {
			return p.applyInclude(node)
		},
	)
}

func (p *Processor) processDirectives(node reflect.Value) (reflect.Value, error) {
	return p.processNode(node,
		func(node reflect.Value) (reflect.Value, error) {
			return p.applyDirectives(node)
		},
	)
}

func (p *Processor) processNode(node reflect.Value, f processFunc) (reflect.Value, error) {
	node = strip(node)
	node, err := f(node)

	if err != nil {
		return reflect.Value{}, err
	}

	nodeKind := node.Kind()

	switch nodeKind {
	case reflect.Map, reflect.Slice:
		ptrAddr := node.Pointer()

		if _, ok := p.seenNodes[ptrAddr]; ok {
			return node, nil
		}

		p.seenNodes[ptrAddr] = struct{}{}
	}

	switch nodeKind {
	case reflect.Map:
		err = p.processMap(node, f)
	case reflect.Slice:
		err = p.processSlice(node, f)
	}

	if err != nil {
		return reflect.Value{}, err
	}

	return node, nil
}

func (p *Processor) processMap(m reflect.Value, f processFunc) error {
	for _, key := range m.MapKeys() {
		keyStr := key.Interface().(string)
		p.keyStack.Push(keyStr)

		node := m.MapIndex(key)
		node, err := p.processNode(node, f)

		if err != nil {
			return err
		}

		m.SetMapIndex(key, node)

		p.keyStack.Pop()
	}

	return nil
}

func (p *Processor) processSlice(s reflect.Value, f processFunc) error {
	sliceLen := s.Len()

	for i := 0; i < sliceLen; i++ {
		idxStr := strconv.Itoa(i)
		p.keyStack.Push(idxStr)

		node := s.Index(i)
		node, err := p.processNode(node, f)

		if err != nil {
			return err
		}

		s.Index(i).Set(node)

		p.keyStack.Pop()
	}

	return nil
}

func (p *Processor) applyInclude(node reflect.Value) (reflect.Value, error) {
	switch node.Kind() {
	case reflect.Map:
		if locators := node.MapIndex(includeKey); locators.IsValid() {
			var err error
			node, err = p.includeSection(locators)

			if err != nil {
				return reflect.Value{}, err
			}
		}
	}

	return node, nil
}

func (p *Processor) applyDirectives(node reflect.Value) (reflect.Value, error) {
	switch node.Kind() {
	case reflect.String:
		str := node.Interface().(string)
		return p.expandRefs(str)
	case reflect.Map:
		if ref := node.MapIndex(refKey); ref.IsValid() {
			return p.resolveRef(ref)
		} else {
			if names := node.MapIndex(underlayKey); names.IsValid() {
				var err error
				node, err = p.mergeLayers(underlayKey, node, names)

				if err != nil {
					return reflect.Value{}, err
				}
			}

			if names := node.MapIndex(overlayKey); names.IsValid() {
				var err error
				node, err = p.mergeLayers(overlayKey, node, names)

				if err != nil {
					return reflect.Value{}, err
				}
			}
		}
	}

	return node, nil
}

func (p *Processor) includeSection(locators reflect.Value) (reflect.Value, error) {
	locators = strip(locators)
	var locList []any
	locsKind := locators.Kind()

	switch locsKind {
	case reflect.String:
		locList = []any{locators.Interface()}
	case reflect.Slice:
		locsLen := locators.Len()
		locList = make([]any, locsLen)

		if locsLen == 0 {
			return reflect.Value{}, fmt.Errorf("%s: at least one configuration locator "+
				"must be sepcified in directive %s at node: %s", errPref, includeKey,
				p.keyStack)
		}

		for i := 0; i < locsLen; i++ {
			loc := locators.Index(i)
			loc = strip(loc)
			locKind := loc.Kind()

			if locKind != reflect.String {
				return reflect.Value{}, fmt.Errorf("%s: configuration locator in %s "+
					"directive must be a string, but got \"%s\" at node: %s", errPref,
					includeKey, locKind, p.keyStack)
			}

			locList[i] = loc.Interface()
		}
	default:
		return reflect.Value{}, fmt.Errorf("%s: value of %s directive must be a "+
			"string or string list, but got \"%s\" at node: %s", errPref, includeKey,
			locsKind, p.keyStack)
	}

	layers, err := p.load(locList)

	if err != nil {
		return reflect.Value{}, err
	}

	var config any

	for _, layer := range layers {
		config = merger.Merge(config, layer)
	}

	return reflect.ValueOf(config), nil
}

func (p *Processor) resolveRef(ref reflect.Value) (reflect.Value, error) {
	ref = strip(ref)
	refKind := ref.Kind()

	switch refKind {
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
				return reflect.Value{}, fmt.Errorf("%s: parameter name in %s directive "+
					"must be a string, but got \"%s\" at node: %s", errPref, refKey,
					nameKind, p.keyStack)
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
				return reflect.Value{}, fmt.Errorf("%s: '\"%s\" parameter in %s directive "+
					"must be a string list, but got \"%s\" at node: %s", errPref,
					firstDefinedKey, refKey, namesKind, p.keyStack)
			}

			namesLen := names.Len()

			for i := 0; i < namesLen; i++ {
				name := names.Index(i)
				name = strip(name)
				nameKind := name.Kind()

				if nameKind != reflect.String {
					return reflect.Value{}, fmt.Errorf("%s: parameter name in \"%s\" parameter "+
						"must be a string, but got \"%s\" at node: %s", errPref, firstDefinedKey,
						nameKind, p.keyStack)
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
	default:
		return reflect.Value{}, fmt.Errorf("%s: value of %s directive must be a "+
			"string or a map, but got \"%s\" at node: %s", errPref, refKey,
			refKind, p.keyStack)
	}

	return reflect.Value{}, nil
}

func (p *Processor) mergeLayers(directiveKey reflect.Value, node reflect.Value,
	names reflect.Value) (reflect.Value, error) {

	names = strip(names)
	var nameList []string
	namesKind := names.Kind()

	switch namesKind {
	case reflect.String:
		name := names.Interface().(string)
		nameList = []string{name}
	case reflect.Slice:
		namesLen := names.Len()
		nameList = make([]string, namesLen)

		if namesLen == 0 {
			return reflect.Value{}, fmt.Errorf("%s: at least one parameter name must "+
				"be sepcified in directive %s at node: %s", errPref, directiveKey,
				p.keyStack)
		}

		for i := 0; i < namesLen; i++ {
			name := names.Index(i)
			name = strip(name)
			nameKind := name.Kind()

			if nameKind != reflect.String {
				return reflect.Value{},
					fmt.Errorf("%s: parameter name in %s directive must be a string,"+
						" but got \"%s\" at node: %s", errPref, directiveKey, nameKind,
						p.keyStack)
			}

			nameList[i] = name.Interface().(string)
		}
	default:
		return reflect.Value{}, fmt.Errorf("%s: value of %s directive must be a "+
			"string or string list, but got \"%s\" at node: %s", errPref, directiveKey,
			namesKind, p.keyStack)
	}

	var layers []any

	for _, name := range nameList {
		layer, err := p.resolveNode(name)

		if err != nil {
			return reflect.Value{}, err
		} else if !layer.IsValid() {
			continue
		}

		layers = append(layers, layer.Interface())
	}

	node.SetMapIndex(directiveKey, reflect.Value{})

	if directiveKey.Equal(underlayKey) {
		layers = append(layers, node.Interface())
	} else {
		layers = append([]any{node.Interface()}, layers...)
	}

	var configSec any

	for _, layer := range layers {
		configSec = merger.Merge(configSec, layer)
	}

	return reflect.ValueOf(configSec), nil
}

func (p *Processor) expandRefs(str string) (reflect.Value, error) {
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
									return reflect.Value{}, err
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

	return reflect.ValueOf(res), nil
}

func (p *Processor) fetchNode(name string) (reflect.Value, error) {
	if node, ok := p.refs[name]; ok {
		return node, nil
	}

	node, err := p.resolveNode(name)

	if err != nil {
		return reflect.Value{}, err
	}

	p.refs[name] = node

	return node, nil
}

func (p *Processor) resolveNode(name string) (reflect.Value, error) {
	stackTemp := p.keyStack
	p.keyStack = newKeyStack(0, keyStackCap)

	defer func() {
		p.keyStack = stackTemp
	}()

	node := p.root
	keys := strings.Split(name, refNameSep)
	keysLen := len(keys)

	for i, keyStr := range keys {
		node = strip(node)
		keyStr = strings.Trim(keyStr, " ")

		switch node.Kind() {
		case reflect.Map:
			key := reflect.ValueOf(keyStr)
			p.keyStack.Push(keyStr)

			child := node.MapIndex(key)

			if !child.IsValid() {
				return reflect.Value{}, nil
			}

			if i == keysLen-1 {
				var err error
				child, err = p.processDirectives(child)

				if err != nil {
					return reflect.Value{}, err
				}

				node.SetMapIndex(key, child)

				return child, nil
			}

			node = child
		case reflect.Slice:
			j, err := strconv.Atoi(keyStr)
			p.keyStack.Push(keyStr)

			if err != nil {
				return reflect.Value{}, fmt.Errorf("%s: invalid array index: %s at node: %s",
					errPref, keyStr, p.keyStack)
			} else if j < 0 || j >= node.Len() {
				return reflect.Value{}, fmt.Errorf("%s: array index out of range: %d at node: %s",
					errPref, j, p.keyStack)
			}

			child := node.Index(j)

			if !child.IsValid() {
				return reflect.Value{}, nil
			}

			if i == keysLen-1 {
				var err error
				child, err = p.processDirectives(child)

				if err != nil {
					return reflect.Value{}, err
				}

				node.Index(j).Set(child)

				return child, nil
			}

			node = child
		default:
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

func (p *Processor) beforeProcess() {
	p.keyStack = newKeyStack(0, keyStackCap)
	p.seenNodes = make(map[uintptr]struct{})
	p.refs = make(map[string]reflect.Value)
}

func (p *Processor) afterProcess() {
	p.keyStack = nil
	p.seenNodes = nil
	p.refs = nil
	p.root = reflect.Value{}
}

func newKeyStack(size, cap int) *keyStack {
	return &keyStack{
		s: make([]string, size, cap),
	}
}

func (s *keyStack) Push(str string) {
	s.s = append(s.s, str)
}

func (s *keyStack) Pop() string {
	l := len(s.s) - 1
	el := s.s[l]
	s.s = s.s[:l]

	return el
}

func (s *keyStack) String() string {
	if len(s.s) > 0 {
		return strings.Join(s.s, refNameSep)
	}

	return "root"
}
