package tidyProviders

import (
	"log"
)

type TidyProvider interface {
	TidyURL(string) (string, error)
	GetURLMatch(string) (bool, error)
}

type TidyProviderInstance struct {
	TidyProvider
}

type TidyProviderInitializer func(map[string]string) (*TidyProviderInstance, error)

type TidyProviders = map[string]*TidyProviderInstance

var Tidiers = make(TidyProviders)

func RegisterTidyProvider(name string, initer TidyProviderInitializer) {
	t, err := initer(nil)
	if err != nil {
		log.Fatalf("Cannot register tidy provider %q multiple times", name)
	}
	Tidiers[name] = t
}

/*
func createTidyProvider(rType string) (TidyProvider, error) {

	t, ok := TidyProvider.InitProvider()
	if !ok {
		return nil, fmt.Errorf("No such tidy provider: %q", rType)
	}

	return t
}
*/

type Null struct {
	tidier TidyProviderInstance
}

var TidyProviderNull TidyProviderInstance

// GetTidyMatch gets an array of regex match strings
func (n *Null) GetURLMatch(s string) (bool, error) {
	return false, nil
}

// initTidy initializes a tidy provider for use
func InitProvider(_ map[string]string) (*TidyProviderInstance, error) {
	return &TidyProviderNull, nil
}

// TidyURL performs the actual tidying of the URL
func (n *Null) TidyURL(s string) (string, error) {
	return s, nil
}
