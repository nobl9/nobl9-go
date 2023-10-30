package validation

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringNotEmpty(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := StringNotEmpty().Validate("                s")
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		err := StringNotEmpty().Validate("     ")
		assert.Error(t, err)
		assert.True(t, HasErrorCode(err, ErrorCodeStringNotEmpty))
	})
}

func TestStringMatchRegexp(t *testing.T) {
	re := regexp.MustCompile("[ab]+")
	t.Run("passes", func(t *testing.T) {
		err := StringMatchRegexp(re).Validate("ab")
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		err := StringMatchRegexp(re).Validate("cd")
		assert.EqualError(t, err, "string does not match regular expresion: [ab]+")
		assert.True(t, HasErrorCode(err, ErrorCodeStringMatchRegexp))
	})
}

func TestStringDenyRegexp(t *testing.T) {
	re := regexp.MustCompile("[ab]+")
	t.Run("passes", func(t *testing.T) {
		err := StringDenyRegexp(re).Validate("cd")
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		err := StringDenyRegexp(re).Validate("ab")
		assert.EqualError(t, err, "string must not match regular expresion: [ab]+")
		assert.True(t, HasErrorCode(err, ErrorCodeStringDenyRegexp))
	})
}

func TestStringIsDNSSubdomain(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for _, input := range []string{
			"test",
			"s",
			"test-this",
			"test-1-this",
			"test1-this",
			"123",
			strings.Repeat("l", 63),
		} {
			err := StringIsDNSSubdomain().Validate(input)
			assert.NoError(t, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		for _, input := range []string{
			"tesT",
			"",
			strings.Repeat("l", 64),
			"test?",
			"test this",
			"1_2",
			"LOL",
		} {
			err := StringIsDNSSubdomain().Validate(input)
			assert.Error(t, err)
			for _, e := range err.(ruleSetError) {
				assert.True(t, HasErrorCode(e, ErrorCodeStringIsDNSSubdomain))
			}
		}
	})
}

func TestStringDescription(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := StringDescription().Validate(strings.Repeat("l", 1050))
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		err := StringDescription().Validate(strings.Repeat("l", 1051))
		assert.Error(t, err)
		assert.True(t, HasErrorCode(err, ErrorCodeStringDescription))
	})
}

func TestStringIsURL(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for _, input := range []string{
			"http://foo.bar#com",
			"http://foobar.com",
			"https://foobar.com",
			"http://foobar.coffee/",
			"http://foobar.中文网/",
			"http://foobar.org/",
			"http://foobar.org:8080/",
			"ftp://foobar.ua/",
			"http://user:pass@www.foobar.com/",
			"http://127.0.0.1/",
			"http://duckduckgo.com/?q=%2F",
			"http://localhost:3000/",
			"http://foobar.com/?foo=bar#baz=qux",
			"http://foobar.com?foo=bar",
			"http://www.xn--froschgrn-x9a.net/",
			"xyz://foobar.com",
			"rtmp://foobar.com",
			"http://www.foo_bar.com/",
			"http://localhost:3000/",
			"http://foobar.com/#baz",
			"http://foobar.com#baz=qux",
			"http://foobar.com/t$-_.+!*\\'(),",
			"http://www.foobar.com/~foobar",
			"http://www.-foobar.com/",
			"http://www.foo---bar.com/",
			"mailto:someone@example.com",
			"irc://irc.server.org/channel",
			"irc://#channel@network",
		} {
			err := StringIsURL().Validate(input)
			assert.NoError(t, err)
		}
	})

	t.Run("fails", func(t *testing.T) {
		for _, input := range []string{
			"foobar.com",
			"",
			"invalid.",
			".com",
			"/abs/test/dir",
			"./rel/test/dir",
			"irc:",
			"http://",
		} {
			err := StringIsURL().Validate(input)
			assert.Error(t, err)
			assert.True(t, HasErrorCode(err, ErrorCodeStringURL))
		}
	})
}
