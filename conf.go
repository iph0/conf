package conf

import (
	"fmt"
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
	config      ProcessorConfig
	root        any
	breadcrumbs []string
	refs        map[string]any
	seen        map[any]struct{}
}

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
	Load(*Locator) (any, error)
}

// M type is a convenient alias for a map[string]any map.
type M = map[string]any

// S type is a convenient alias for a []any slice.
type S = []any

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

// Load method loads configuration tree using configuration locators.
// Configuration locator can be a string or a map of type map[string]any.
// Map type can be used to specify default configuration layers. The merge
// priority of loaded configuration layers depends on the order of configuration
// locators. Layers loaded by rightmost locator have highest priority.
func (p *Processor) Load(locators ...any) (M, error) {
	if len(locators) == 0 {
		panic(fmt.Errorf("%s: no configuration locators specified", errPref))
	}

	iConfig, err := p.load(locators)

	if err != nil {
		return nil, err
	}

	if iConfig == nil {
		return nil, nil
	}

	if !p.config.DisableProcessing {
		iConfig, err = p.process(iConfig)
	}

	if err != nil {
		return nil, err
	}

	config, ok := iConfig.(M)

	if !ok {
		return nil, fmt.Errorf("%s: loaded configuration has invalid type %T",
			errPref, config)
	}

	return config, nil
}

func (p *Processor) load(locators []any) (any, error) {
	var config any

	for _, iRawLoc := range locators {
		switch rawLoc := iRawLoc.(type) {
		case M:
			config = merger.Merge(config, rawLoc)
		case string:
			loc, err := ParseLocator(rawLoc)

			if err != nil {
				return nil, err
			}

			loader, ok := p.config.Loaders[loc.Loader]

			if !ok {
				return nil, fmt.Errorf("%s: loader not found for configuration locator %s",
					errPref, loc)
			}

			layer, err := loader.Load(loc)

			if err != nil {
				return nil, err
			}

			if layer == nil {
				continue
			}

			config = merger.Merge(config, layer)
		default:
			return nil, fmt.Errorf("%s: invalid type of configuration locator: %T",
				errPref, rawLoc)
		}
	}

	return config, nil
}

func (p *Processor) process(config any) (any, error) {
	p.breadcrumbs = make([]string, 0, 10)
	p.refs = make(map[string]any)
	p.seen = make(map[any]struct{})

	defer func() {
		p.root = nil
		p.breadcrumbs = nil
		p.refs = nil
		p.seen = nil
	}()

	config, err := p.processValue(config)

	if err != nil {
		return nil, err
	}

	p.root = config
	err = p.processNode(p.root)

	if err != nil {
		return nil, fmt.Errorf("%s at %s", err, p.processContext())
	}

	return config, nil
}

func (p *Processor) processNode(node any) error {
	switch n := node.(type) {
	case map[string]any:
		return p.processMap(n)
	case []any:
		return p.processSlice(n)
	}

	return nil
}

func (p *Processor) processMap(m map[string]any) error {
	addrStr := fmt.Sprintf("%p", m)

	if _, ok := p.seen[addrStr]; ok {
		return nil
	}

	p.seen[addrStr] = struct{}{}

	for key, value := range m {
		p.pushCrumb(key)
		value, err := p.processValue(value)

		if err != nil {
			return err
		}

		err = p.processNode(value)

		if err != nil {
			return err
		}

		m[key] = value
		p.popCrumb()
	}

	return nil
}

func (p *Processor) processSlice(s []any) error {
	addrStr := fmt.Sprintf("%p", s)

	if _, ok := p.seen[addrStr]; ok {
		return nil
	}

	p.seen[addrStr] = struct{}{}

	for i, value := range s {
		indexStr := strconv.Itoa(i)
		p.pushCrumb(indexStr)
		value, err := p.processValue(value)

		if err != nil {
			return err
		}

		err = p.processNode(value)

		if err != nil {
			return err
		}

		s[i] = value
		p.popCrumb()
	}

	return nil
}

func (p *Processor) processValue(value any) (any, error) {
	switch v := value.(type) {
	case string:
		return p.expandRefs(v)
	case map[string]any:
		if ref, ok := v["$ref"]; ok {
			return p.processRef(ref)
		} else if locators, ok := v["$include"]; ok {
			return p.processInclude(locators)
		}
	}

	return value, nil
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
								value, err := p.resolveRef(name)

								if err != nil {
									return "", err
								}

								if value != nil {
									res += fmt.Sprintf("%v", value)
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

func (p *Processor) processRef(ref any) (any, error) {
	switch r := ref.(type) {
	case string:
		return p.resolveRef(r)
	case map[string]any:
		if nameIf, ok := r["name"]; ok {
			name, ok := nameIf.(string)

			if !ok {
				return nil, fmt.Errorf("%s: invalid type of reference name: %T", errPref,
					nameIf)
			}

			node, err := p.resolveRef(name)

			if err != nil {
				return nil, err
			} else if node != nil {
				return node, nil
			}
		} else if namesIf, ok := r["firstOf"]; ok {
			names, ok := namesIf.([]any)

			if !ok {
				return nil, fmt.Errorf("%s: invalid type of \"firstOf\" field: %T",
					errPref, namesIf)
			}

			for _, nameIf := range names {
				name, ok := nameIf.(string)

				if !ok {
					return nil, fmt.Errorf("%s: invalid type of reference name: %T", errPref,
						nameIf)
				}

				node, err := p.resolveRef(name)

				if err != nil {
					return nil, err
				} else if node != nil {
					return node, nil
				}
			}
		}

		if node, ok := r["default"]; ok {
			return node, nil
		}
	default:
		return nil, fmt.Errorf("%s: invalid type of $ref directive: %T", errPref, ref)
	}

	return nil, nil
}

func (p *Processor) processInclude(locators any) (any, error) {
	locatorList, ok := locators.([]any)

	if !ok {
		return nil, fmt.Errorf("%s: invalid type of $include directive: %T",
			errPref, locators)
	}

	branch, err := p.load(locatorList)

	if err != nil {
		return nil, err
	}

	return branch, nil
}

func (p *Processor) resolveRef(name string) (any, error) {
	value, ok := p.refs[name]

	if ok {
		return value, nil
	}

	value, err := p.findNode(name)

	if err != nil {
		return nil, err
	}

	p.refs[name] = value

	return value, nil
}

func (p *Processor) findNode(name string) (any, error) {
	currNode := p.root
	tokens := strings.Split(name, refNameSep)

	for i, tkn := range tokens {
		tkn = strings.Trim(tkn, " ")

		switch n := currNode.(type) {
		case map[string]any:
			childNode, ok := n[tkn]

			if !ok {
				return nil, nil
			}

			crumbs := p.breadcrumbs
			p.breadcrumbs = tokens[:i+1]

			childNode, err := p.processValue(childNode)

			if err != nil {
				return nil, err
			}

			n[tkn] = childNode
			currNode = childNode

			p.breadcrumbs = crumbs
		case []any:
			j, err := strconv.Atoi(tkn)

			if err != nil {
				return nil, fmt.Errorf("%s: invalid slice index", errPref)
			} else if j < 0 || j >= len(n) {
				return nil, fmt.Errorf("%s: slice index out of range", errPref)
			}

			crumbs := p.breadcrumbs
			p.breadcrumbs = tokens[:i+1]

			childNode, err := p.processValue(n[j])

			if err != nil {
				return nil, err
			}

			n[j] = childNode
			currNode = childNode

			p.breadcrumbs = crumbs
		default:
			return nil, nil
		}

		if currNode == nil {
			return nil, nil
		}
	}

	return currNode, nil
}

func (p *Processor) pushCrumb(bc string) {
	p.breadcrumbs = append(p.breadcrumbs, bc)
}

func (p *Processor) popCrumb() {
	p.breadcrumbs = p.breadcrumbs[:len(p.breadcrumbs)-1]
}

func (p *Processor) processContext() string {
	return strings.Join(p.breadcrumbs, refNameSep)
}
