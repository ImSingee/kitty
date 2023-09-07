package config

import (
	"os"

	"github.com/ImSingee/go-ex/ee"
)

var ErrNotExist = os.ErrNotExist

func IsNotExist(err error) bool {
	return ee.Is(err, ErrNotExist)
}
