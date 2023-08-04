package manifest

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestFilterByKind(t *testing.T) {
	t.Run("nil objects slice", func(t *testing.T) {
		objects := FilterByKind[customObject](nil)
		assert.Nil(t, objects)
	})

	t.Run("empty objects slice", func(t *testing.T) {
		objects := FilterByKind[customObject]([]Object{})
		assert.Nil(t, objects)
	})

	t.Run("no matching objects", func(t *testing.T) {
		objects := FilterByKind[customObject]([]Object{
			customProjectScopedObject{},
			customProjectScopedObject{},
		})
		assert.Nil(t, objects)
	})

	t.Run("different objects", func(t *testing.T) {
		objects := FilterByKind[customObject]([]Object{
			customObject{},
			customProjectScopedObject{},
			customObject{},
			customProjectScopedObject{},
		})
		assert.Len(t, objects, 2)
		assert.IsType(t, []customObject{}, objects)
	})
}

func TestValidate(t *testing.T) {
	t.Run("nil objects slice", func(t *testing.T) {
		err := Validate(nil)
		assert.NoError(t, err)
	})

	t.Run("empty objects slice", func(t *testing.T) {
		err := Validate([]Object{})
		assert.NoError(t, err)
	})

	t.Run("no errors", func(t *testing.T) {
		err := Validate([]Object{
			customObject{validationError: nil},
			customObject{validationError: nil},
		})
		assert.NoError(t, err)
	})

	t.Run("errors", func(t *testing.T) {
		err := Validate([]Object{
			customObject{validationError: nil},
			customObject{validationError: errors.New("I failed!")},
			customObject{validationError: errors.New("I failed too!")},
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "I failed!\nI failed too!")
	})
}

func TestSetDefaultProject(t *testing.T) {
	for name, test := range map[string]struct {
		Input    []Object
		Expected []Object
	}{
		"nil objects slice": {
			Input:    nil,
			Expected: nil,
		},
		"empty objects slice": {
			Input:    []Object{},
			Expected: []Object{},
		},
		"different objects": {
			Input: []Object{
				customProjectScopedObject{project: ""},
				customObject{},
				customProjectScopedObject{project: "this"},
				customProjectScopedObject{project: ""},
			},
			Expected: []Object{
				customProjectScopedObject{project: "default"},
				customObject{},
				customProjectScopedObject{project: "this"},
				customProjectScopedObject{project: "default"},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			objects := SetDefaultProject(test.Input, "default")
			assert.Equal(t, test.Expected, objects)
		})
	}
}

type customObject struct {
	validationError error
}

func (c customObject) GetVersion() string { return "" }

func (c customObject) GetKind() Kind { return 0 }

func (c customObject) GetName() string { return "" }

func (c customObject) Validate() error { return c.validationError }

type customProjectScopedObject struct {
	customObject
	project string
}

func (c customProjectScopedObject) GetProject() string { return c.project }

func (c customProjectScopedObject) SetProject(project string) Object { c.project = project; return c }
