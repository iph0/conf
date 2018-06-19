package conf

import (
	"fmt"
	"strings"
)

type locator struct {
	Provider    string
	bareLocator string
}

func parseLocator(rawLoc string) (*locator, error) {
	if rawLoc == "" {
		return nil, fmt.Errorf("%s: empty configuration locator specified", errPref)
	}

	locTokens := strings.SplitN(rawLoc, ":", 2)

	if len(locTokens) < 2 || locTokens[0] == "" {
		return nil, fmt.Errorf("%s: missing provider name in configuration locator",
			errPref)
	}

	return &locator{
		Provider:    locTokens[0],
		bareLocator: locTokens[1],
	}, nil
}

func (l *locator) String() string {
	return fmt.Sprintf("%s:%s", l.Provider, l.bareLocator)
}
