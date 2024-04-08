package conf

import (
	"fmt"
	"strings"
)

// Locator is used by configuration processor and loaders to load configuration layers.
type Locator struct {
	Loader string
	Value  string
}

// ParseLocator method creates Locator instance from the string.
func ParseLocator(locStr string) (*Locator, error) {
	if locStr == "" {
		return nil, fmt.Errorf("%s: empty configuration locator specified", errPref)
	}

	tokens := strings.SplitN(locStr, ":", 2)

	if len(tokens) < 2 || tokens[0] == "" {
		return nil, fmt.Errorf("%s: missing loader name in configuration locator",
			errPref)
	}

	return &Locator{
		Loader: tokens[0],
		Value:  tokens[1],
	}, nil
}

// String method convert Locator instance to the string.
func (l *Locator) String() string {
	return fmt.Sprintf("%s:%s", l.Loader, l.Value)
}
