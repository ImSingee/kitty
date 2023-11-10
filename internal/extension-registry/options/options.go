package eroptions

import (
	"fmt"

	"github.com/ysmood/gson"
)

type AnyOptions = map[string]gson.JSON // will parse on-demand

func AsAnyOptions(s map[string]any) AnyOptions {
	return gson.New(s).Map()
}

// AsAnyOptionsOrDollar will convert the given value to AnyOptions.
//
// If o is map[string]any, it will return AsAnyOptions(o),
// or else, it will return AsAnyOptions(map[string]any{"$": o})
func AsAnyOptionsOrDollar(o any) AnyOptions {
	if o == nil {
		return nil
	}

	if o, ok := o.(map[string]any); ok {
		return AsAnyOptions(o)
	}

	return wrappedOptions(o)
}

func wrappedOptions(v any) AnyOptions {
	return AsAnyOptions(map[string]any{"$": v})
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

// RenameDollarKey
// rename o["$"] to o[to]
//
// force: if true, will overwrite the existing o[to]
// return: same object of o
func RenameDollarKey(o AnyOptions, to string, force bool) AnyOptions {
	if o == nil {
		return nil
	}

	_, toExists := o[to]
	if !toExists || force {
		_, dollarExists := o["$"]
		if dollarExists {
			o[to] = o["$"]
			delete(o, "$")
		}
	}

	return o
}

// Exists return whether the key exists in o
func Exists(o AnyOptions, key string) bool {
	_, exists := o[key]
	return exists
}

// Get key from o
//
// if the key is not found, return nil
func Get(o AnyOptions, key string) AnyOptions {
	switch v := o[key].Val().(type) {
	case nil:
		return nil
	case map[string]any:
		return AsAnyOptions(v)
	default:
		return wrappedOptions(v)
	}
}

// Merge two AnyOptions, return a new merged object
//
// override: if there's same key in o and with, whether to override it (make o[key] = with[key])
func Merge(o AnyOptions, with AnyOptions, override bool) AnyOptions {
	result := make(AnyOptions)
	for k, v := range o {
		result[k] = v
	}

	if override {
		for k, v := range with {
			result[k] = v
		}
	} else {
		for k, v := range with {
			if _, exists := result[k]; !exists {
				result[k] = v
			}
		}
	}

	return result
}
