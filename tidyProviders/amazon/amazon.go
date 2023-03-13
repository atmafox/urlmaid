package amazon

import (
	"fmt"
	"regexp"

	"github.com/atmafox/urlmaid/tidyProviders"
)

var TidyProviderAmazon tidyProviders.TidyProviderInstance

type amazonTidyProvider struct {
	match []string
}

var amazon = amazonTidyProvider{
	match: []string{`(?P<domain>amazon.com)`, `(?P<domain>www.amazon.com)`},
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
	for i := range amazon.match {
		r := regexp.MustCompile(amazon.match[i])

		if b := r.MatchString(s); b == true {
			return b, nil
		}
	}

	return false, nil
}

func (c amazonTidyProvider) TidyURL(s string) (string, error) {
	ru := regexp.MustCompile(`(?P<useful>/dp/[[:alnum:]]+)/`)

	var m []string
	var d string

	for r := range amazon.match {
		rt := regexp.MustCompile(amazon.match[r])

		m = rt.FindStringSubmatch(s)
		if m != nil {
			d = m[rt.SubexpIndex("domain")]
		}
	}

	m = ru.FindStringSubmatch(s)

	u := m[ru.SubexpIndex("useful")]

	out := fmt.Sprintf("https://%s%s", d, u)
	return out, nil
}
