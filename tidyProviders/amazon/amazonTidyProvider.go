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
	match: []string{`(?P<domain>amazon.com\)`, `(?P<domain>www.amazon.com)`},
}

func init() {
	tidyProviders.RegisterTidyProvider("AMAZON", &initAmazon)
}

func initAmazon(_ map[string]string) (tidyProviders.TidyProviderInstance, error) {
	// No config to do
	return &TidyProviderAmazon, nil
}

func (c *amazonTidyProvider) GetURLMatch(s string) (bool, error) {
	for i := range amazon.match {
		r, err := regexp.Compile(amazon.match[i])
		if err != nil {
			return false, err
		}

		if b := r.MatchString(s); b == true {
			return b, nil
		}
	}

	return false, nil
}

func (c *amazonTidyProvider) TidyURL(s string) (string, error) {
	ru, err := regexp.Compile(`(?P<useful>/dp/[[:alnum:]]+)/`)
	if err != nil {
		return "", err
	}

	var d string

	for r := range amazon.match {
		rt, err := regexp.Compile(amazon.match[r])
		if err != nil {
			return "", err
		}

		d = rt.FindString(s)
	}

	u := ru.FindString(s)

	out := fmt.Sprintf("https://%s%s", d, u)
	return out, nil
}
