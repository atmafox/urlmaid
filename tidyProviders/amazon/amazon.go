package amazon

import (
	"regexp"

	"github.com/atmafox/urlmaid/tidyProviders"
)

var TidyProviderAmazon tidyProviders.TidyProviderInstance

type amazonTidyProvider struct {
	match     []string
	transform []string
}

var amazon = amazonTidyProvider{
	match: []string{`(?P<domain>amazon.com)`, `(?P<domain>www.amazon.com)`},
	transform: []string{`(?P<useful>/dp/[^/?]+)`,
		`(?P<useful>/deal/[^/?]+)`,
		`(?P<useful>/discover/bn/[^/?]+)`,
		`(?P<useful>/gp/goldbox/[^/?]+)`},
}

func init() {
	tidyProviders.RegisterTidyProvider("amazon", initAmazon)
}

func initAmazon(_ map[string]string) (*tidyProviders.TidyProviderInstance, error) {
	// No config to do
	TidyProviderAmazon.TidyProvider = amazon
	return &TidyProviderAmazon, nil
}

func (c amazonTidyProvider) GetURLMatch(s string) (bool, error) {
	m, _, err := c.GetRegexps()
	if err != nil {
		return false, err
	}

	for i := range m {
		r := regexp.MustCompile(m[i])

		if b := r.MatchString(s); b == true {
			return b, nil
		}
	}

	return false, nil
}

func (c amazonTidyProvider) TidyURL(s string) (string, error) {
	m, t, err := c.GetRegexps()
	if err != nil {
		return "", err
	}

	out, err := tidyProviders.ProcessRegexps(s, m, t)
	if err != nil {
		return "", err
	}

	return out, nil
}

func (c amazonTidyProvider) GetRegexps() ([]string, []string, error) {
	return amazon.match, amazon.transform, nil
}
