package overlaybd

const (
	Root          = "/opt/overlaybd"
	ServiceBinary = Root + "/bin/overlaybd-service"

	ConverterRoot        = Root + "/convert"
	ConverterBinary      = ConverterRoot + "/bin/turboOCI-apply"
	ConverterMergeBinary = ConverterRoot + "/bin/overlaybd-convert-acs"

	BaseLayer = Root + "/baselayers/.commit"
)

// overlaybd consts
const (
	// Naa prefix for loopback devices in configfs
	// for example snID 128, the loopback device config in /sys/kernel/config/target/loopback/naa.1990000000000128
	obdLoopNaaPrefix = 199

	// hba number used to create tcmu devices in configfs
	// all overlaybd devices are configured in /sys/kernel/config/target/core/user_999999999/
	// devices ares identified by their snID /sys/kernel/config/target/core/user_999999999/dev_$snID
	obdHbaNum = 999999999

	// param used to restrict tcmu devices mmap memory size for iSCSI data.
	// it is worked by setting max_data_area_mb for devices in configfs.
	obdMaxDataAreaMB = 4

	timeout = 20

	maxAttachAttempts = 400
)
