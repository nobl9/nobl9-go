package validation

import (
	"net/url"

	"github.com/pkg/errors"
)

func URL() SingleRule[url.URL] {
	return NewSingleRule(func(v url.URL) error { return validateURL(v) }).WithErrorCode(ErrorCodeURL)
}

func validateURL(v url.URL) error {
	if v.Scheme == "" {
		return errors.New("valid URL must have a scheme (e.g. https://)")
	}
	if v.Host == "" && v.Fragment == "" && v.Opaque == "" {
		return errors.New("valid URL must contain either host, fragment or opaque data")
	}
	return nil
}
