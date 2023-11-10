package binkey

import "runtime"

type BinKey string

const (
	LinuxAmd64  BinKey = "linux-amd64"
	LinuxArm64  BinKey = "linux-arm64"
	DarwinAmd64 BinKey = "darwin-amd64"
	DarwinArm64 BinKey = "darwin-arm64"
)

func GetCurrentBinKey() BinKey {
	return BinKey(runtime.GOOS + "-" + runtime.GOARCH)
}
