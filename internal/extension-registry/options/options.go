package eroptions

import (
	"fmt"

	"github.com/ysmood/gson"
)

type AnyOptions = map[string]gson.JSON // will parse on-demand

func AsAnyOptions(s map[string]any) AnyOptions {
	return gson.New(s).Map()
}

// AsAnyOptionsOrKey will convert the given value to AnyOptions.
//
// If o is map[string]any, it will return AsAnyOptions(o),
// or else, it will return AsAnyOptions(map[string]any{key: o})
func AsAnyOptionsOrKey(o any, key string) AnyOptions {
	if o == nil {
		return nil
	}

	if o, ok := o.(map[string]any); ok {
		return AsAnyOptions(o)
	}

	return AsAnyOptions(map[string]any{
		key: o,
	})
}

// AsAnyOptionsOrError will convert the given value to AnyOptions.
//
// if o is not map[string]any, it will return
// AsAnyOptions(map[string]any{"error": "invalid or malformed value"})
func AsAnyOptionsOrError(o any) AnyOptions {
	if o == nil {
		return nil
	}

	if o, ok := o.(map[string]any); ok {
		return AsAnyOptions(o)
	}

	return AsAnyOptions(map[string]any{
		"error": "invalid or malformed value",
	})
}

// CastToError will cast the given AnyOptions to error.
//
// Note: this function won't treat nil options as an error.
func CastToError(o AnyOptions) error {
	if o == nil {
		return nil
	}

	if err, ok := o["error"]; ok {
		return fmt.Errorf("%v", err)
	}

	return nil
}
