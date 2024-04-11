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
	config    ProcessorConfig
	keyStack  *keyStack
	seenNodes map[uintptr]struct{}
	refs      map[string]reflect.Value
	root      reflect.Value
}

var (
	refKey          = reflect.ValueOf("$ref")
	includeKey      = reflect.ValueOf("$include")
	underlayKey     = reflect.ValueOf("$underlay")
	overlayKey      = reflect.ValueOf("$overlay")
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
			var err error
			layer, err = p.processIncludes(layer)

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
		config, err = p.processDirectives(config)

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

func (p *Processor) processIncludes(layer any) (any, error) {
	p.beforeProcess()
	defer p.afterProcess()

	lyr := reflect.ValueOf(layer)

	lyr, err := p.processNode(lyr,
		func(node reflect.Value) (reflect.Value, error) {
			return p.applyInclude(node)
		},
	)

	if err != nil {
		return nil, err
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

func (p *Processor) processDirectives(config any) (any, error) {
	p.beforeProcess()
	defer p.afterProcess()

	conf := reflect.ValueOf(config)
	p.root = conf

	conf, err := p.processNode(conf,
		func(node reflect.Value) (reflect.Value, error) {
			return p.applyDirectives(node)
		},
	)

	if err != nil {
		return nil, err
	}

	return conf.Interface(), nil
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
			node, err = p.include(locators)

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
			return p.applyOverlays(node)
		}
	}

	return node, nil
}

func (p *Processor) applyOverlays(node reflect.Value) (reflect.Value, error) {
	if refs := node.MapIndex(underlayKey); refs.IsValid() {
		var err error
		node, err = p.underlay(node, refs)

		if err != nil {
			return reflect.Value{}, err
		}
	}

	if refs := node.MapIndex(overlayKey); refs.IsValid() {
		var err error
		node, err = p.overlay(node, refs)

		if err != nil {
			return reflect.Value{}, err
		}
	}

	return node, nil
}

func (p *Processor) include(locators reflect.Value) (reflect.Value, error) {
	locators = strip(locators)
	var locList []any

	switch locators.Kind() {
	case reflect.String:
		locList = []any{locators.Interface()}
	case reflect.Slice:
		locsLen := locators.Len()
		locList = make([]any, locsLen)

		for i := 0; i < locsLen; i++ {
			loc := locators.Index(i)
			loc = strip(loc)
			locKind := loc.Kind()

			if locKind != reflect.String {
				return reflect.Value{},
					fmt.Errorf("%s: configuration locator in %s directive must be a string,"+
						" but got: %s at %s", errPref, includeKey, locKind, p.keyStack)
			}

			locList[i] = loc.Interface()
		}
	}

	if locList == nil {
		return reflect.Value{}, fmt.Errorf("%s: malformed directive: %s at %s",
			errPref, includeKey, p.keyStack)
	}

	layers, err := p.load(locList)

	if err != nil {
		return reflect.Value{}, err
	}

	config := p.merge(layers)

	return reflect.ValueOf(config), nil
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
					fmt.Errorf("%s: reference name must be a string, but got: %s at %s",
						errPref, nameKind, p.keyStack)
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
					fmt.Errorf("%s: '\"%s\" must be an array, but got: %s at %s", errPref,
						firstDefinedKey, namesKind, p.keyStack)
			}

			namesLen := names.Len()

			for i := 0; i < namesLen; i++ {
				name := names.Index(i)
				name = strip(name)
				nameKind := name.Kind()

				if nameKind != reflect.String {
					return reflect.Value{},
						fmt.Errorf("%s: reference name in \"%s\" must be a string, but got: %s at %s",
							errPref, firstDefinedKey, nameKind, p.keyStack)
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

	return reflect.Value{}, fmt.Errorf("%s: malformed directive: %s at %s",
		errPref, refKey, p.keyStack)
}

func (p *Processor) underlay(node reflect.Value, refs reflect.Value) (reflect.Value, error) {
	refs = strip(refs)
	var refList []string

	switch refs.Kind() {
	case reflect.String:
		ref := refs.Interface().(string)
		refList = []string{ref}
	case reflect.Slice:
		refsLen := refs.Len()
		refList = make([]string, refsLen)

		for i := 0; i < refsLen; i++ {
			ref := refs.Index(i)
			ref = strip(ref)
			refKind := ref.Kind()

			if refKind != reflect.String {
				return reflect.Value{},
					fmt.Errorf("%s: reference name in %s directive must be a string,"+
						" but got: %s at %s", errPref, includeKey, refKind, p.keyStack)
			}

			refList[i] = ref.Interface().(string)
		}
	}

	if refList == nil {
		return reflect.Value{}, fmt.Errorf("%s: malformed directive: %s at %s",
			errPref, underlayKey, p.keyStack)
	}

	var layers []any

	for _, ref := range refList {
		layer, err := p.findNode(ref)

		if err != nil {
			return reflect.Value{}, err
		}

		layers = append(layers, layer.Interface())
	}

	layers = append(layers, node.Interface())
	mergedNode := p.merge(layers)

	return reflect.ValueOf(mergedNode), nil
}

func (p *Processor) overlay(node reflect.Value, refs reflect.Value) (reflect.Value, error) {
	refs = strip(refs)
	var refList []string

	switch refs.Kind() {
	case reflect.String:
		ref := refs.Interface().(string)
		refList = []string{ref}
	case reflect.Slice:
		refsLen := refs.Len()
		refList = make([]string, refsLen)

		for i := 0; i < refsLen; i++ {
			ref := refs.Index(i)
			ref = strip(ref)
			refKind := ref.Kind()

			if refKind != reflect.String {
				return reflect.Value{},
					fmt.Errorf("%s: reference name in %s directive must be a string,"+
						" but got: %s at %s", errPref, includeKey, refKind, p.keyStack)
			}

			refList[i] = ref.Interface().(string)
		}
	}

	if refList == nil {
		return reflect.Value{}, fmt.Errorf("%s: malformed directive: %s at %s",
			errPref, overlayKey, p.keyStack)
	}

	var layers []any

	for _, ref := range refList {
		layer, err := p.findNode(ref)

		if err != nil {
			return reflect.Value{}, err
		}

		layers = append(layers, layer.Interface())
	}

	layers = append([]any{node.Interface()}, layers...)
	mergedNode := p.merge(layers)

	return reflect.ValueOf(mergedNode), nil
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
	stackTemp := p.keyStack
	p.keyStack = newKeyStack(0, 10)

	defer func() {
		p.keyStack = stackTemp
	}()

	node := p.root
	tokens := strings.Split(name, refNameSep)
	tokensLen := len(tokens)

	for i := 0; i < tokensLen; i++ {
		node = strip(node)
		tokens[i] = strings.Trim(tokens[i], " ")

		switch node.Kind() {
		case reflect.Map:
			key := reflect.ValueOf(tokens[i])
			p.keyStack.Push(tokens[i])

			child := node.MapIndex(key)
			node = child
		case reflect.Slice:
			j, err := strconv.Atoi(tokens[i])
			p.keyStack.Push(tokens[i])

			if err != nil {
				return reflect.Value{}, fmt.Errorf("%s: invalid array index: %s at %s",
					errPref, tokens[i], p.keyStack)
			} else if j < 0 || j >= node.Len() {
				return reflect.Value{}, fmt.Errorf("%s: array index out of range: %d at %s",
					errPref, j, p.keyStack)
			}

			child := node.Index(j)
			node = child
		default:
			return reflect.Value{}, nil
		}

		if !node.IsValid() {
			return reflect.Value{}, nil
		}
	}

	var err error
	node, err = p.processNode(node,
		func(node reflect.Value) (reflect.Value, error) {
			return p.applyDirectives(node)
		},
	)

	if err != nil {
		return reflect.Value{}, err
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
	p.keyStack = newKeyStack(0, 10)
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
	return strings.Join(s.s, refNameSep)
}
