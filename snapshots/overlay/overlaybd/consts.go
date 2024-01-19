package overlaybd

const (
	Root          = "/opt/overlaybd"
	ServiceBinary = Root + "/bin/overlaybd"

	ConverterRoot        = Root + "/convert"
	ConverterBinary      = ConverterRoot + "/bin/turboOCI-apply"
	ConverterMergeBinary = ConverterRoot + "/bin/overlaybd-convert-acs"
)
