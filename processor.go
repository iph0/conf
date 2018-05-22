package conf

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Processor type represents procesor instance.
type Processor struct {
	config   reflect.Value
	path     []string
	varIndex map[string]reflect.Value
}

// Process method walks through the configuration tree and expands all variables.
func (p *Processor) Process(config map[string]interface{}) {
	p.config = reflect.ValueOf(config)
	p.path = make([]string, 0, 10)
	p.varIndex = make(map[string]reflect.Value)

	p.walk(p.config)
}

func (p *Processor) walk(node reflect.Value) {
	nodeKind := node.Kind()

	if nodeKind == reflect.Interface {
		node = node.Elem()
		nodeKind = node.Kind()
	}

	if nodeKind == reflect.Map {
		for _, key := range node.MapKeys() {
			p.pushPathSegment(key.Interface().(string))

			value := p.process(node.MapIndex(key))
			node.SetMapIndex(key, value)
			p.walk(value)

			p.popPathSegment()
		}
	} else if nodeKind == reflect.Slice {
		sliceLen := node.Len()

		for i := 0; i < sliceLen; i++ {
			p.pushPathSegment(strconv.Itoa(i))

			value := node.Index(i)
			value.Set(p.process(value))
			p.walk(value)

			p.popPathSegment()
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

	return value
}

func (p *Processor) expandVars(src string) string {
	var result string

	srcRunes := []rune(src)
	srcLen := len(srcRunes)
	i, j := 0, 0

	for j < srcLen {
		if srcRunes[j] == '$' && j+1 < srcLen {
			if srcRunes[j+1] == '{' {
				result += string(srcRunes[i:j])

				for i, j = j, j+2; j < srcLen; j++ {
					if srcRunes[j] == '}' {
						name := string(srcRunes[i+2 : j])
						value := p.resolveVar(name)
						result += fmt.Sprintf("%v", value.Interface())
						i, j = j+1, j+1

						break
					}
				}

				continue
			}

			if srcRunes[j+1] == '$' &&
				j+2 < srcLen &&
				srcRunes[j+2] == '{' {

				result += string(srcRunes[i:j])

				for i, j = j, j+3; j < srcLen; j++ {
					if srcRunes[j] == '}' {
						result += string(srcRunes[i+1 : j+1])
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
	value, ok := p.varIndex[name]

	if ok {
		return value
	}

	nameTokens := strings.Split(name, ".")

	if nameTokens[0] == "" {
		nameLen := len(nameTokens)
		pathLen := len(p.path)
		var offset int

		for i := 0; i < nameLen; i++ {
			if nameTokens[i] != "" {
				nameTokens = nameTokens[i:]
				offset = i

				break
			}
		}

		if offset < pathLen {
			temp := make([]string, 0, pathLen+nameLen)
			temp = append(temp, p.path[:pathLen-offset]...)
			nameTokens = append(temp, nameTokens...)
		}
	}

	value = p.fetchValue(nameTokens)
	name = strings.Join(nameTokens, ".")
	p.varIndex[name] = value

	return value
}

func (p *Processor) fetchValue(name []string) reflect.Value {
	value := p.config
	nameLen := len(name)

	for i := 0; i < nameLen; i++ {
		valKind := value.Kind()

		if valKind == reflect.Interface {
			value = value.Elem()
			valKind = value.Kind()
		}

		if valKind == reflect.Map {
			key := reflect.ValueOf(name[i])
			value = value.MapIndex(key)
		} else if valKind == reflect.Slice {
			j, err := strconv.Atoi(name[i])

			if err != nil {
				// TODO
			}

			value = value.Index(j)
		}

		if !value.IsValid() {
			return reflect.ValueOf("")
		}
	}

	return value
}

func (p *Processor) pushPathSegment(bc string) {
	p.path = append(p.path, bc)
}

func (p *Processor) popPathSegment() {
	p.path = p.path[:len(p.path)-1]
}
