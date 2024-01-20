package overlaybd

import (
	"fmt"
	"strings"
)

func overlaybdLoopbackDeviceID(id string) string {
	paddings := strings.Repeat("0", 13-len(id))
	return fmt.Sprintf("naa.%d%s%s", obdLoopNaaPrefix, paddings, id)
}

func overlaybdLoopbackDeviceLunPath(id string) string {
	return fmt.Sprintf("/sys/kernel/config/target/loopback/%s/tpgt_1/lun/lun_0", id)
}

func overlaybdLoopbackDevicePath(id string) string {
	return fmt.Sprintf("/sys/kernel/config/target/loopback/%s", id)
}

func overlaybdTargetPath(id string) string {
	return fmt.Sprintf("/sys/kernel/config/target/core/user_%d/dev_%s", obdHbaNum, id)
}

func scsiBlockDevicePath(deviceNumber string) string {
	return fmt.Sprintf("/sys/class/scsi_device/%s:0/device/block", deviceNumber)
}
