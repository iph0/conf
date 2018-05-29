package conf

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const nameSep = "."

var (
	nilVal   = reflect.ValueOf(nil)
	emptyStr = reflect.ValueOf("")
)

// Processor type represents procesor instance.
type Processor struct {
	root        reflect.Value
	breadcrumbs []string
	varIndex    map[string]reflect.Value
	seenNodes   map[reflect.Value]bool
}

// Process method walks through the configuration tree and expands all variables
// in string values.
func (p *Processor) Process(root interface{}) error {
	p.root = reflect.ValueOf(root)
	p.breadcrumbs = make([]string, 0, 10)
	p.varIndex = make(map[string]reflect.Value)
	p.seenNodes = make(map[reflect.Value]bool)

	err := p.walk(p.root)

	if err != nil {
		return err
	}

	return nil
}

func (p *Processor) walk(node reflect.Value) error {
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
	p.seenNodes[node] = true

	if nodeKind == reflect.Map {
		for _, key := range node.MapKeys() {
			p.pushCrumb(key.Interface().(string))

			value, err := p.process(node.MapIndex(key))

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
			p.pushCrumb(strconv.Itoa(i))

			value := node.Index(i)
			pValue, err := p.process(value)

			if err != nil {
				return err
			}

			value.Set(pValue)
			err = p.walk(value)

			if err != nil {
				return err
			}

			p.popCrumb()
		}
	}

	return nil
}

func (p *Processor) process(value reflect.Value) (reflect.Value, error) {
	valKind := value.Kind()

	if valKind == reflect.Interface {
		value = value.Elem()
		valKind = value.Kind()
	}

	if valKind == reflect.String {
		valueStr := value.Interface().(string)
		valueStr, err := p.expandVars(valueStr)

		if err != nil {
			return nilVal, err
		}

		return reflect.ValueOf(valueStr), nil
	}

	if valKind == reflect.Map {
		key := reflect.ValueOf("@var")
		name := value.MapIndex(key)

		if name.IsValid() {
			nameKind := name.Kind()

			if nameKind == reflect.Interface {
				name = name.Elem()
				nameKind = name.Kind()
			}

			if nameKind != reflect.String {
				return nilVal, fmt.Errorf("%s: invalid @var directive: %s",
					errPref, strings.Join(p.breadcrumbs, nameSep))
			}

			nameStr := name.Interface().(string)
			value, err := p.resolveVar(nameStr)

			if err != nil {
				return nilVal, err
			}

			return value, nil
		}
	}

	return value, nil
}

func (p *Processor) expandVars(src string) (string, error) {
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

func (p *Processor) resolveVar(name string) (reflect.Value, error) {
	if name == "" {
		return p.root, nil
	}

	tokens := strings.Split(name, nameSep)

	if tokens[0] == "" {
		tokens = p.expandName(tokens)
		name = strings.Join(tokens, nameSep)
	}

	value, ok := p.varIndex[name]

	if ok {
		return value, nil
	}

	value, err := p.fetchVal(tokens)

	if err != nil {
		return nilVal, err
	}

	p.varIndex[name] = value

	return value, nil
}

func (p *Processor) expandName(name []string) []string {
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

func (p *Processor) fetchVal(name []string) (reflect.Value, error) {
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
				return nilVal, fmt.Errorf("%s: invalid slice index: %s",
					errPref, strings.Join(name, nameSep))
			} else if i < 0 || i >= node.Len() {
				return nilVal, fmt.Errorf("%s: slice index out of range: %s",
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
		value, err = p.process(value)

		if err != nil {
			return nilVal, err
		}

		key := reflect.ValueOf(name[len(name)-1])
		node.SetMapIndex(key, value)
	} else if nodeKind == reflect.Slice {
		pValue, err := p.process(value)

		if err != nil {
			return nilVal, err
		}

		value.Set(pValue)
	}

	p.breadcrumbs = crumbs

	return value, nil
}

func (p *Processor) pushCrumb(bc string) {
	p.breadcrumbs = append(p.breadcrumbs, bc)
}

func (p *Processor) popCrumb() {
	p.breadcrumbs = p.breadcrumbs[:len(p.breadcrumbs)-1]
}
