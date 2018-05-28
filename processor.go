package conf

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const nameSep = "."

var emptyStr = reflect.ValueOf("")

// Processor type represents procesor instance.
type Processor struct {
	root        reflect.Value
	breadcrumbs []string
	varIndex    map[string]reflect.Value
	seenNodes   map[reflect.Value]bool
}

// Process method walks through the configuration tree and expands all variables
// in string values.
func (p *Processor) Process(root interface{}) {
	p.root = reflect.ValueOf(root)
	p.breadcrumbs = make([]string, 0, 10)
	p.varIndex = make(map[string]reflect.Value)
	p.seenNodes = make(map[reflect.Value]bool)

	p.walk(p.root)
}

func (p *Processor) walk(node reflect.Value) {
	nodeKind := node.Kind()

	if nodeKind == reflect.Interface {
		node = node.Elem()
		nodeKind = node.Kind()
	}

	if nodeKind != reflect.Map &&
		nodeKind != reflect.Slice {

		return
	}

	if _, ok := p.seenNodes[node]; ok {
		return
	}
	p.seenNodes[node] = true

	if nodeKind == reflect.Map {
		for _, key := range node.MapKeys() {
			p.pushCrumb(key.Interface().(string))

			value := p.process(node.MapIndex(key))
			node.SetMapIndex(key, value)
			p.walk(value)

			p.popCrumb()
		}
	} else { // Slice
		sliceLen := node.Len()

		for i := 0; i < sliceLen; i++ {
			p.pushCrumb(strconv.Itoa(i))

			value := node.Index(i)
			value.Set(p.process(value))
			p.walk(value)

			p.popCrumb()
		}
	}
}

func (p *Processor) process(value reflect.Value) reflect.Value {
	valKind := value.Kind()

	if valKind == reflect.Interface {
		value = value.Elem()
		valKind = value.Kind()
	}

	if valKind == reflect.String {
		str := value.Interface().(string)
		str = p.expandVars(str)

		return reflect.ValueOf(str)
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

			if nameKind == reflect.String {
				return p.resolveVar(name.Interface().(string))
			}
		}
	}

	return value
}

func (p *Processor) expandVars(src string) string {
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
							value := p.resolveVar(name)
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

	return result
}

func (p *Processor) resolveVar(name string) reflect.Value {
	if name == "" {
		return p.root
	}

	tokens := strings.Split(name, nameSep)

	if tokens[0] == "" {
		tokens = p.expandName(tokens)
		name = strings.Join(tokens, nameSep)
	}

	value, ok := p.varIndex[name]

	if ok {
		return value
	}

	value = p.findVal(tokens)
	p.varIndex[name] = value

	return value
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

func (p *Processor) findVal(name []string) reflect.Value {
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

			if err != nil ||
				i >= node.Len() {

				return emptyStr
			}

			value = node.Index(i)
		} else {
			return emptyStr
		}

		if !value.IsValid() {
			return emptyStr
		}
	}

	crumbs := p.breadcrumbs
	p.breadcrumbs = name
	nodeKind := node.Kind()

	if nodeKind == reflect.Map {
		key := reflect.ValueOf(name[len(name)-1])
		value = p.process(value)
		node.SetMapIndex(key, value)
	} else if nodeKind == reflect.Slice {
		value.Set(p.process(value))
	}

	p.breadcrumbs = crumbs

	return value
}

func (p *Processor) pushCrumb(bc string) {
	p.breadcrumbs = append(p.breadcrumbs, bc)
}

func (p *Processor) popCrumb() {
	p.breadcrumbs = p.breadcrumbs[:len(p.breadcrumbs)-1]
}
