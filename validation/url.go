package validation

import (
	"net/url"

	"github.com/pkg/errors"
)

func URL() SingleRule[*url.URL] {
	return NewSingleRule(func(v *url.URL) error { return validateURL(v) }).WithErrorCode(ErrorCodeURL)
}

func validateURL(u *url.URL) error {
	if u.Scheme == "" {
		return errors.New("valid URL must have a scheme (e.g. https://)")
	}
	if u.Host == "" && u.Fragment == "" && u.Opaque == "" {
		return errors.New("valid URL must contain either host, fragment or opaque data")
	}
	return nil
}
