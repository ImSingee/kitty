package extregistry

// BinOptions specify where to download a binary
//
// {[osKey]:
type BinOptions AnyOptions

func parseBinOptions(in any) BinOptions {
	return asAnyOptionsOrError(in)
}

func (o BinOptions) ToDownloader() Installer {
	return &distInstaller{o}
}
