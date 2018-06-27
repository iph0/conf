package conf

import (
	"fmt"
	"strings"
)

// Locator is used by configuration loader and providers to load configuration
// layers.
type Locator struct {
	Provider    string
	BareLocator string
}

// ParseLocator method creates Locator instance from the string.
func ParseLocator(rawLoc string) (*Locator, error) {
	if rawLoc == "" {
		return nil, fmt.Errorf("%s: empty configuration locator specified", errPref)
	}

	locTokens := strings.SplitN(rawLoc, ":", 2)

	if len(locTokens) < 2 || locTokens[0] == "" {
		return nil, fmt.Errorf("%s: missing provider name in configuration locator",
			errPref)
	}

	return &Locator{
		Provider:    locTokens[0],
		BareLocator: locTokens[1],
	}, nil
}

// String method convert Locator instance to the string.
func (l *Locator) String() string {
	return fmt.Sprintf("%s:%s", l.Provider, l.BareLocator)
}
