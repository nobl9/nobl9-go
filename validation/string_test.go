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
		assert.EqualError(t, err, "string does not match regular expresion: '[ab]+'")
		assert.True(t, HasErrorCode(err, ErrorCodeStringMatchRegexp))
	})
	t.Run("examples output", func(t *testing.T) {
		err := StringMatchRegexp(re, "ab", "a", "b").Validate("cd")
		assert.EqualError(t, err, "string does not match regular expresion: '[ab]+' (e.g. 'ab', 'a', 'b')")
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
		assert.EqualError(t, err, "string must not match regular expresion: '[ab]+'")
		assert.True(t, HasErrorCode(err, ErrorCodeStringDenyRegexp))
	})
	t.Run("examples output", func(t *testing.T) {
		err := StringDenyRegexp(re, "ab", "a", "b").Validate("ab")
		assert.EqualError(t, err, "string must not match regular expresion: '[ab]+' (e.g. 'ab', 'a', 'b')")
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

func TestStringASCII(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for _, input := range []string{
			"foobar",
			"0987654321",
			"test@example.com",
			"1234abcDEF",
			"",
		} {
			err := StringASCII().Validate(input)
			assert.NoError(t, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		for _, input := range []string{
			"ｆｏｏbar",
			"ｘｙｚ０９８",
			"１２３456",
			"ｶﾀｶﾅ",
		} {
			err := StringASCII().Validate(input)
			assert.Error(t, err)
			assert.True(t, HasErrorCode(err, ErrorCodeStringASCII))
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
			err := StringURL().Validate(input)
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
			err := StringURL().Validate(input)
			assert.Error(t, err)
			assert.True(t, HasErrorCode(err, ErrorCodeStringURL))
		}
	})
}

func TestStringJSON(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := StringJSON().Validate(`{"foo": "bar"}`)
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		err := StringJSON().Validate(`{]}`)
		assert.Error(t, err)
		assert.True(t, HasErrorCode(err, ErrorCodeStringJSON))
	})
}

func TestStringContains(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := StringContains("th", "is").Validate("this")
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		err := StringContains("th", "ht").Validate("one")
		assert.Error(t, err)
		assert.EqualError(t, err, "string must contain the following substrings: 'th', 'ht'")
		assert.True(t, HasErrorCode(err, ErrorCodeStringContains))
	})
}
