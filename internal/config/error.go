package config

import (
	"github.com/ImSingee/go-ex/ee"
	"os"
)

var ErrNotExist = os.ErrNotExist

func IsNotExist(err error) bool {
	return ee.Is(err, ErrNotExist)
}
