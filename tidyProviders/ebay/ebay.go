package ebay

import (
	"fmt"
	"regexp"

	"github.com/atmafox/urlmaid/tidyProviders"
)

var TidyProviderEBay tidyProviders.TidyProviderInstance

type ebayTidyProvider struct {
	match []string
}

var ebay = ebayTidyProvider{
	match: []string{`(?P<domain>ebay.com)`, `(?P<domain>www.ebay.com)`},
}

func init() {
	tidyProviders.RegisterTidyProvider("ebay", initEBay)
}

func initEBay(_ map[string]string) (*tidyProviders.TidyProviderInstance, error) {
	// No config to do
	TidyProviderEBay.TidyProvider = ebay
	return &TidyProviderEBay, nil
}

func (c ebayTidyProvider) GetURLMatch(s string) (bool, error) {
	for i := range ebay.match {
		r := regexp.MustCompile(ebay.match[i])

		if b := r.MatchString(s); b == true {
			return b, nil
		}
	}

	return false, nil
}

func (c ebayTidyProvider) TidyURL(s string) (string, error) {
	ru := regexp.MustCompile(`(?P<useful>/itm/[[:digit:]]+)\?`)

	var m []string
	var d string

	for r := range ebay.match {
		rt := regexp.MustCompile(ebay.match[r])

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
