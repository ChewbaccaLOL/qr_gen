package cli

import (
	"fmt"
	"strconv"
)

type OptionalString struct {
	Value string
	IsSet bool
}

type OptionalInt struct {
	Value int
	IsSet bool
}

type OptionalFloat struct {
	Value float64
	IsSet bool
}

type OptionalBool struct {
	Value bool
	IsSet bool
}

func (o *OptionalString) String() string {
	return o.Value
}

func (o *OptionalString) SetValue(value string) {
	o.Value = value
	o.IsSet = true
}

func (o *OptionalString) Set(value string) error {
	o.SetValue(value)
	return nil
}

func (o *OptionalInt) String() string {
	return strconv.Itoa(o.Value)
}

func (o *OptionalInt) SetValue(value int) {
	o.Value = value
	o.IsSet = true
}

func (o *OptionalInt) Set(value string) error {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid int: %w", err)
	}
	o.SetValue(parsed)
	return nil
}

func (o *OptionalFloat) String() string {
	return strconv.FormatFloat(o.Value, 'f', -1, 64)
}

func (o *OptionalFloat) SetValue(value float64) {
	o.Value = value
	o.IsSet = true
}

func (o *OptionalFloat) Set(value string) error {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid float: %w", err)
	}
	o.SetValue(parsed)
	return nil
}

func (o *OptionalBool) String() string {
	return strconv.FormatBool(o.Value)
}

func (o *OptionalBool) IsBoolFlag() bool {
	return true
}

func (o *OptionalBool) SetValue(value bool) {
	o.Value = value
	o.IsSet = true
}

func (o *OptionalBool) Set(value string) error {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid bool: %w", err)
	}
	o.SetValue(parsed)
	return nil
}
