package extregistry

type gotInstaller struct {
	o GoInstallOptions
}

func (d *gotInstaller) Install(o *InstallOptions) error {
	//if err := d.o["error"].Val(); err != nil {
	//	return fmt.Errorf("%v", err)
	//}
	//
	//osKey := GetCurrentBinKey()
	//url := d.o[osKey].Val()
	//
	//url := d.o["url"].Val()
	//if url == nil {
	//	return ErrInstallerNotApplicable
	//}
	//
	//urlString, ok := url.(string)
	//if !ok {
	//	return ee.New("`url` key in bin options is not string")
	//}
	//
	//return downloadFileTo(urlString, o.To, 0755, o.ShowProgress)
}
