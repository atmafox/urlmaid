package main

import (
	"fmt"
	"log"
)

type TidyProvider interface {
	models.TidyProvider
}

type TidyProviderInitializer func(map[string]string) (TidyProvider, error)

var TidyProviders = map[string]TidyProviderInitializer{}

func RegisterTidyProvider(name string, init TidyProviderInitializer) {
	if _, ok := TidyProviders[name]; ok {
		log.Fatalf("Cannot register tidy provider %q multiple times", name)
	}
	TidyProviderTypes[name] = init
}

func createTidyProvider(rType string) (TidyProvider, error) {

	initer, ok := TidyProvider[rType]
	if !ok {
		return nil, fmt.Errorf("No such tidy provider: %q", rType)
	}

	return initer()
}

type Null struct{}

func (n Null) GetTidyProviderDomains() ([]string, error) {
	return nil, nil
}

func init() {
	RegisterRegistrarType("NULL", func(map[string]string) (TidyProvider, error) {
		return Null{}, nil
	})
}

func (n Null) TidyURL(s string) (string, error) {
	return s, nil
}
