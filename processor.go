package conf

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const nameSep = "."

var (
	varKey     = reflect.ValueOf("@var")
	includeKey = reflect.ValueOf("@include")
	zeroVal    = reflect.ValueOf(nil)
	emptyStr   = reflect.ValueOf("")
)

type processor struct {
	loader      *Loader
	root        reflect.Value
	breadcrumbs []string
	varIndex    map[string]reflect.Value
	seenNodes   map[reflect.Value]struct{}
}

func newProcessor(loader *Loader) *processor {
	return &processor{
		loader: loader,
	}
}

func (p *processor) Process(root interface{}) error {
	p.root = reflect.ValueOf(root)
	p.breadcrumbs = make([]string, 0, 10)
	p.varIndex = make(map[string]reflect.Value)
	p.seenNodes = make(map[reflect.Value]struct{})

	err := p.walk(p.root)

	if err != nil {
		return err
	}

	return nil
}

func (p *processor) walk(node reflect.Value) error {
	nodeKind := node.Kind()

	if nodeKind == reflect.Interface {
		node = node.Elem()
		nodeKind = node.Kind()
	}

	if nodeKind != reflect.Map &&
		nodeKind != reflect.Slice {

		return nil
	}

	if _, ok := p.seenNodes[node]; ok {
		return nil
	}

	p.seenNodes[node] = struct{}{}

	if nodeKind == reflect.Map {
		for _, key := range node.MapKeys() {
			keyStr := key.Interface().(string)
			p.pushCrumb(keyStr)

			value := node.MapIndex(key)
			value, err := p.processNode(value)

			if err != nil {
				return err
			}

			node.SetMapIndex(key, value)
			err = p.walk(value)

			if err != nil {
				return err
			}

			p.popCrumb()
		}
	} else { // Slice
		sliceLen := node.Len()

		for i := 0; i < sliceLen; i++ {
			indexStr := strconv.Itoa(i)
			p.pushCrumb(indexStr)

			value := node.Index(i)
			val, err := p.processNode(value)

			if err != nil {
				return err
			}

			value.Set(val)
			err = p.walk(value)

			if err != nil {
				return err
			}

			p.popCrumb()
		}
	}

	return nil
}

func (p *processor) processNode(node reflect.Value) (reflect.Value, error) {
	nodeKind := node.Kind()

	if nodeKind == reflect.Interface {
		node = node.Elem()
		nodeKind = node.Kind()
	}

	if nodeKind == reflect.String {
		str := node.Interface().(string)
		str, err := p.expandVarsString(str)

		if err != nil {
			return zeroVal, err
		}

		return reflect.ValueOf(str), nil
	}

	if nodeKind == reflect.Map {
		if name := node.MapIndex(varKey); name.IsValid() {
			node, err := p.expandVar(name)

			if err != nil {
				return zeroVal, err
			}

			return node, nil
		}

		if pat := node.MapIndex(includeKey); pat.IsValid() {
			node, err := p.include(pat)

			if err != nil {
				return zeroVal, err
			}

			return node, nil
		}
	}

	return node, nil
}

func (p *processor) expandVar(name reflect.Value) (reflect.Value, error) {
	nameKind := name.Kind()

	if nameKind == reflect.Interface {
		name = name.Elem()
		nameKind = name.Kind()
	}

	if nameKind != reflect.String {
		return zeroVal, fmt.Errorf("%s: invalid @var directive: %s",
			errPref, strings.Join(p.breadcrumbs, nameSep))
	}

	nameStr := name.Interface().(string)
	value, err := p.resolveVar(nameStr)

	if err != nil {
		return zeroVal, err
	}

	return value, nil
}

func (p *processor) include(pattern reflect.Value) (reflect.Value, error) {
	patternKind := pattern.Kind()

	if patternKind == reflect.Interface {
		pattern = pattern.Elem()
		patternKind = pattern.Kind()
	}

	if patternKind != reflect.String {
		return zeroVal, fmt.Errorf("%s: invalid @include directive: %s",
			errPref, strings.Join(p.breadcrumbs, nameSep))
	}

	patternStr := pattern.Interface().(string)
	data, err := p.loader.Load(patternStr)

	if err != nil {
		return zeroVal, err
	}

	return reflect.ValueOf(data), nil
}

func (p *processor) expandVarsString(src string) (string, error) {
	var result string

	srcRunes := []rune(src)
	srcLen := len(srcRunes)
	i, j := 0, 0

	for j < srcLen {
		if srcRunes[j] == '$' && j+1 < srcLen {
			var esc bool
			k := 1

			if srcRunes[j+1] == '$' {
				esc = true
				k++
			}

			if srcRunes[j+k] == '{' {
				result += string(srcRunes[i:j])

				for i, j = j, j+k+1; j < srcLen; j++ {
					if srcRunes[j] == '}' {
						if esc {
							result += string(srcRunes[i+1 : j+1])
						} else {
							name := string(srcRunes[i+2 : j])
							value, err := p.resolveVar(name)

							if err != nil {
								return "", err
							}

							result += fmt.Sprintf("%v", value.Interface())
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

	result += string(srcRunes[i:j])

	return result, nil
}

func (p *processor) resolveVar(name string) (reflect.Value, error) {
	if name == "" {
		return p.root, nil
	}

	tokens := strings.Split(name, nameSep)

	if tokens[0] == "" {
		tokens = p.expandName(tokens)
		name = strings.Join(tokens, nameSep)
	}

	if value, ok := p.varIndex[name]; ok {
		return value, nil
	}

	value, err := p.fetchValue(tokens)

	if err != nil {
		return zeroVal, err
	}

	p.varIndex[name] = value

	return value, nil
}

func (p *processor) expandName(name []string) []string {
	nameLen := len(name)
	crumbsLen := len(p.breadcrumbs)
	i := 0

	for ; i < nameLen; i++ {
		if name[i] != "" {
			break
		}
	}

	if i == nameLen {
		i--

		if i >= crumbsLen {
			return p.breadcrumbs[:0]
		}

		return p.breadcrumbs[:crumbsLen-i]
	}

	if i >= crumbsLen {
		return name[i:]
	}

	return append(
		append([]string{}, p.breadcrumbs[:crumbsLen-i]...),
		name[i:]...,
	)
}

func (p *processor) fetchValue(name []string) (reflect.Value, error) {
	var node reflect.Value
	value := p.root

	for _, token := range name {
		token := strings.Trim(token, " ")
		valKind := value.Kind()

		if valKind == reflect.Interface {
			value = value.Elem()
			valKind = value.Kind()
		}

		if valKind == reflect.Map {
			node = value
			key := reflect.ValueOf(token)
			value = node.MapIndex(key)
		} else if valKind == reflect.Slice {
			node = value
			i, err := strconv.Atoi(token)

			if err != nil {
				return zeroVal, fmt.Errorf("%s: invalid slice index: %s",
					errPref, strings.Join(name, nameSep))
			} else if i < 0 || i >= node.Len() {
				return zeroVal, fmt.Errorf("%s: slice index out of range: %s",
					errPref, strings.Join(name, nameSep))
			}

			value = node.Index(i)
		} else {
			return emptyStr, nil
		}

		if !value.IsValid() {
			return emptyStr, nil
		}
	}

	crumbs := p.breadcrumbs
	p.breadcrumbs = name
	nodeKind := node.Kind()

	if nodeKind == reflect.Map {
		var err error
		value, err = p.processNode(value)

		if err != nil {
			return zeroVal, err
		}

		key := reflect.ValueOf(name[len(name)-1])
		node.SetMapIndex(key, value)
	} else if nodeKind == reflect.Slice {
		val, err := p.processNode(value)

		if err != nil {
			return zeroVal, err
		}

		value.Set(val)
	}

	p.breadcrumbs = crumbs

	return value, nil
}

func (p *processor) pushCrumb(bc string) {
	p.breadcrumbs = append(p.breadcrumbs, bc)
}

func (p *processor) popCrumb() {
	p.breadcrumbs = p.breadcrumbs[:len(p.breadcrumbs)-1]
}
