package lintstaged

import (
	"runtime"
	"sync"
)

var defaultMaxArgLength = sync.OnceValue(func() int {
	switch runtime.GOOS {
	case "darwin":
		return 262144
	case "windows":
		return 8191
	default:
		return 131072
	}
})

func chunkFiles(filenames []string, maxArgLength int) [][]string {
	if maxArgLength <= 0 { // no limit
		return [][]string{filenames}
	}

	arrays := make([][]string, 0, 4)

	currentArray := ([]string)(nil)
	currentLength := 0

	addToArray := func(e string) {
		if currentLength != 0 && currentLength+len(e) > maxArgLength {
			// create new array
			arrays = append(arrays, currentArray)

			currentLength = 0
			currentArray = ([]string)(nil)
		}

		currentArray = append(currentArray, e)
		currentLength += len(e) + 1
	}

	for _, filename := range filenames {
		addToArray(filename)
	}

	if currentArray != nil {
		arrays = append(arrays, currentArray)
	}

	return arrays
}
