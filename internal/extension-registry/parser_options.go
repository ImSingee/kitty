package extregistry

import "github.com/ysmood/gson"

type AnyOptions = map[string]gson.JSON // will parse on-demand

func asAnyOptions(s map[string]any) AnyOptions {
	return gson.New(s).Map()
}

func asAnyOptionsOrKey(o any, key string) AnyOptions {
	if o == nil {
		return nil
	}

	if o, ok := o.(map[string]any); ok {
		return asAnyOptions(o)
	}

	return asAnyOptions(map[string]any{
		key: o,
	})
}

func asAnyOptionsOrError(o any) AnyOptions {
	if o == nil {
		return nil
	}

	if o, ok := o.(map[string]any); ok {
		return asAnyOptions(o)
	}

	return asAnyOptions(map[string]any{
		"error": "invalid or malformed value",
	})
}
