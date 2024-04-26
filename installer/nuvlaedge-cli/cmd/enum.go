package cmd

import "fmt"

type strEnum struct {
	Allowed []string
	Value   string
}

func newStrEnum(defaultValue string, allowed ...string) *strEnum {
	return &strEnum{
		Allowed: allowed,
		Value:   defaultValue,
	}
}

func (e *strEnum) String() string {
	return e.Value
}

func (e *strEnum) Set(value string) error {
	isIncluded := func(values []string, val string) bool {
		for _, v := range values {
			if v == val {
				return true
			}
		}
		return false
	}

	if !isIncluded(e.Allowed, value) {
		return fmt.Errorf("invalid value: %s. Allowed values are: %v", value, e.Allowed)
	}
	e.Value = value
	return nil
}

func (e *strEnum) Type() string {
	return "string"
}
