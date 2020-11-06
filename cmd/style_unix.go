package cmd

func defaultStyleName() string {
	return "unicode"
}

func isFuseFs(m Mount) bool {
	//FIXME: implement
	return false
}

func isNetworkFs(m Mount) bool {
	//FIXME: implement
	return false
}

func isSpecialFs(m Mount) bool {
	return m.Fstype == "devfs"
}
