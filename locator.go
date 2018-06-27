package conf

import (
	"fmt"
	"strings"
)

// Locator TODO
type Locator struct {
	Source      string
	BareLocator string
}

// ParseLocator TODO
func ParseLocator(rawLoc string) (*Locator, error) {
	if rawLoc == "" {
		return nil, fmt.Errorf("%s: empty configuration locator specified", errPref)
	}

	locTokens := strings.SplitN(rawLoc, ":", 2)

	if len(locTokens) < 2 || locTokens[0] == "" {
		return nil, fmt.Errorf("%s: missing source name in configuration locator",
			errPref)
	}

	return &Locator{
		Source:      locTokens[0],
		BareLocator: locTokens[1],
	}, nil
}

func (l *Locator) String() string {
	return fmt.Sprintf("%s:%s", l.Source, l.BareLocator)
}
