package overlaybd

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/containerd/containerd/archive"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/snapshots/overlay/overlaybd/config"
	"github.com/containerd/containerd/snapshots/overlay/roDriver"
)

const (
	lowerFile   = "device"
	tmpTarIndex = "layer.tar.meta"
	dataFile    = ".data_file"  // top layer data file for lsmd
	idxFile     = ".data_index" // top layer index file for lsmd

	labelSnapshotRef = "containerd.io/snapshot.ref"

	// labelKeyOverlayBDBlobDigest is the annotation key in the manifest to
	// describe the digest of blob in OverlayBD format.
	//
	// NOTE: The annotation is part of image layer blob's descriptor.
	labelKeyOverlayBDBlobDigest = "containerd.io/snapshot/overlaybd/blob-digest"

	// labelKeyOverlayBDBlobSize is the annotation key in the manifest to
	// describe the size of blob in OverlayBD format.
	//
	// NOTE: The annotation is part of image layer blob's descriptor.
	labelKeyOverlayBDBlobSize = "containerd.io/snapshot/overlaybd/blob-size"

	// labelTurboOCIDigest is the index annotation key for image layer digest
	labelTurboOCIDigest = "containerd.io/snapshot/overlaybd/turbo-oci/target-digest"

	// labelTurboOCIMediaType is the index annotation key for image layer media type
	labelTurboOCIMediaType = "containerd.io/snapshot/overlaybd/turbo-oci/target-media-type"

	overlaybdCreateBinary = "/opt/overlaybd/bin/overlaybd-create"
)

type Overlaybd struct {
}

func New() (roDriver.RoDriver, error) {
	// ensure overlaybd available
	if err := SupportsOverlaybd(); err != nil {
		return nil, err
	}
	return &Overlaybd{}, nil
}

func SupportsOverlaybd() error {
	// 1. ensure overlaybd-service exist
	_, err := os.Stat(ServiceBinary)
	if err != nil {
		return fmt.Errorf("error stating overlaybd service binary %v: %w", ServiceBinary, err)
	}
	// 2. ensure overlaybd converters exist
	_, err = os.Stat(ConverterBinary)
	if err != nil {
		return fmt.Errorf("error stating overlaybd converter binary %v: %w", ConverterBinary, err)
	}
	_, err = os.Stat(ConverterMergeBinary)
	if err != nil {
		return fmt.Errorf("error stating overlaybd merger binary %v: %w", ConverterMergeBinary, err)
	}
	_, err = os.Stat(BaseLayer)
	if err != nil {
		return fmt.Errorf("error stating overlaybd base layer %v: %w", BaseLayer, err)
	}
	return nil
}

// PreProcess is used for preprocessing layer and judge skip fetch
func (o *Overlaybd) PreProcess(ctx context.Context, keyDir, parentDir, parent string, labels map[string]string) (skipFetch bool, err error) {
	blobSize, hasBDBlobSize := labels[labelKeyOverlayBDBlobSize]
	blobDigest, hasBDBlobDigest := labels[labelKeyOverlayBDBlobDigest]
	skipFetch = false
	if !hasBDBlobSize || !hasBDBlobDigest {
		return false, nil
	}
	if _, ok := labels[labelTurboOCIDigest]; ok {
		// construct turboOCI config
		err := config.ConstructTurboOCISpec(ctx, parent, labels, keyDir, parentDir, blobDigest, blobSize)
		if err != nil {
			return false, err
		}
		return false, nil
	} else {
		// construct overlaybd config
		err := config.ConstructOverlayBDSpec(ctx, parent, labels, keyDir, parentDir, blobDigest, blobSize)
		if err != nil {
			return false, err
		}
		return true, nil
	}
}

// ActiveMount for overlaybd will prepare a device from provided parents, and return corresponding mount
func (o *Overlaybd) ActiveMount(ctx context.Context, snDir string, id string, parentDir string, parentPaths []string, opts ...roDriver.Opt) ([]mount.Mount, error) {
	// 0. if local change turbo oci, invoke convert binary
	localConverted := true
	for i := range parentPaths {
		parentPaths[i] = filepath.Join(parentPaths[i], tmpTarIndex)
		if _, err := os.Stat(parentPaths[i]); err != nil {
			localConverted = false
			break
		}
	}
	if localConverted {
		args := []string{"--workdir", filepath.Join(snDir, "tmp")}
		for _, parentPath := range parentPaths {
			args = append([]string{"--meta", parentPath}, args...)
		}
		start := time.Now()
		log.G(ctx).Debugf("convert merger: merging %v with args : %v", snDir, args)

		// FIXME(kuzhi.zm): 当使用单pod单盘时，convert中做出现invalid cross-device link 的报错，
		// kubelet提供的loop device和lower dir出现跨盘的情况，
		// 当前convert工具中如果出现rename失败的情况下需要做copy的动作，这里需要看如何去优化
		var elapsed int64
		cmd := exec.Command(ConverterMergeBinary, args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			elapsed = time.Now().Sub(start).Milliseconds()
			log.G(ctx).Errorf("failed to merge tar file to overlaybd device with command %v: %v. elapsed: %dms", cmd.String(), string(output), elapsed)
			return nil, fmt.Errorf("failed to merge image file: %w", err)
		}
		elapsed = time.Now().Sub(start).Milliseconds()
		log.G(ctx).Infof("convert merger for %s done. (elapsed: %dms)", snDir, elapsed)
	}
	// 1. overlaybd-create
	if err := overlaybdCreate(snDir); err != nil {
		return nil, fmt.Errorf("overlaybd-create failed: %w", err)
	}
	// 2. create config file
	if err := config.ConstructOverlayBDWritableSpec(snDir, parentDir); err != nil {
		return nil, fmt.Errorf("construct overlaybd writable spec failed: %w", err)
	}
	// 3. do overlaybd process
	dev, err := doOverlayBD(ctx, snDir, id)
	if err != nil {
		return nil, fmt.Errorf("construct overlaybd debug log path failed: %w", err)
	}
	// 4. write device to file
	if err := ioutil.WriteFile(filepath.Join(snDir, lowerFile), []byte(dev), 0666); err != nil {
		log.G(ctx).Errorf("failed to write device path file[%s/%s], dev: %s", snDir, lowerFile, dev)
		return nil, err
	}
	// 5. return device ext4 mount
	return []mount.Mount{
		{
			Source:  dev,
			Type:    "ext4",
			Options: []string{"ro"},
		},
	}, nil
}

func overlaybdCreate(snDir string) error {
	str := fmt.Sprintf("%s -s %s %s 256 ", overlaybdCreateBinary, filepath.Join(snDir, dataFile), filepath.Join(snDir, idxFile))
	cmd := exec.Command("/bin/bash", "-c", str)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("overlaybd-create failed. output: %s: %w", out, err)
	}
	return nil
}

func doOverlayBD(ctx context.Context, snDir string, snID string) (_ string, retErr error) {
	starttime := time.Now()
	targetPath := overlaybdTargetPath(snID)
	err := os.MkdirAll(targetPath, 0700)
	if err != nil {
		return "", fmt.Errorf("failed to create target dir for %s: %w", targetPath, err)
	}

	defer func() {
		if retErr != nil {
			log.G(ctx).Infof("failed tcmuMount, clean, err: %v", retErr)
			rerr := os.RemoveAll(targetPath)
			if rerr != nil {
				log.G(ctx).Warnf("failed to clean target dir %s", targetPath)
			}
		}
	}()

	if err = ioutil.WriteFile(path.Join(targetPath, "control"), ([]byte)(fmt.Sprintf("dev_config=overlaybd/%s;%s", config.OverlaybdConfPath(snDir), snID)), 0666); err != nil {
		return "", fmt.Errorf("failed to write target dev_config for %s: %w", targetPath, err)
	}

	err = ioutil.WriteFile(path.Join(targetPath, "control"), ([]byte)(fmt.Sprintf("max_data_area_mb=%d", obdMaxDataAreaMB)), 0666)
	if err != nil {
		return "", fmt.Errorf("failed to write target max_data_area_mb for %s: %w", targetPath, err)
	}

	elapsed := float64(time.Since(starttime)) / float64(time.Millisecond)
	log.G(ctx).Debugf("to prepre time used: %v", elapsed)

	debugLogPath := config.OverlaybdInitDebuglogPath(snDir)
	log.G(ctx).Debugf("result file: %s", debugLogPath)
	err = os.RemoveAll(debugLogPath)
	if err != nil {
		return "", fmt.Errorf("failed to remote result file for %s: %w", targetPath, err)
	}

	// 5s
	for retry := 0; retry < 100; retry++ {
		err = ioutil.WriteFile(path.Join(targetPath, "enable"), ([]byte)("1"), 0666)
		if err != nil {
			perror, ok := err.(*os.PathError)
			if ok {
				if perror.Err == syscall.EAGAIN {
					log.G(ctx).Infof("write %s returned EAGAIN, retry", targetPath)
					time.Sleep(50 * time.Millisecond)
					continue
				}
			}
			return "", fmt.Errorf("failed to write enable for %s: %w", targetPath, err)
		} else {
			break
		}
	}

	defer func() {
		if retErr != nil {
			log.G(ctx).Infof("failed tcmuMount, kill process %v", snID)
			killProcess(ctx, snID, int(syscall.SIGINT))
		}
	}()

	if err != nil {
		return "", fmt.Errorf("failed to write enable for %s: %w", targetPath, err)
	}
	// fixed by fuweid
	err = ioutil.WriteFile(
		path.Join(targetPath, "attrib", "cmd_time_out"),
		([]byte)(fmt.Sprintf("%v", math.MaxInt32/1000)), 0666)
	if err != nil {
		return "", fmt.Errorf("failed to update cmd_time_out: %w", err)
	}

	// 20s
	// read the init-debug.log for readable
	err = fmt.Errorf("timeout")
	retryTimes := timeout * 1000 / 20 // timeoutInSec / 20ms
	for retry := 0; retry < retryTimes; retry++ {
		data, derr := os.ReadFile(debugLogPath)
		if derr != nil {
			time.Sleep(20 * time.Millisecond)
			err = derr
			continue
		}
		if string(data) == "success" {
			err = nil
			break
		} else {
			return "", fmt.Errorf("failed to enable target for %s: %s", targetPath, data)
		}
	}
	if err != nil {
		log.G(ctx).Warnf("timeout to start device for snID: %s, lastErr: %v", snID, err)
		return "", fmt.Errorf("failed to enable target for %s: %w", targetPath, err)
	}

	elapsed = float64(time.Since(starttime)) / float64(time.Millisecond)
	log.G(ctx).Debugf("to backstore started time used: %v", elapsed)

	loopDevID := overlaybdLoopbackDeviceID(snID)
	loopDevPath := overlaybdLoopbackDevicePath(loopDevID)

	err = os.MkdirAll(loopDevPath, 0700)
	if err != nil {
		return "", fmt.Errorf("failed to create loopback dir %s: %w", loopDevPath, err)
	}

	tpgtPath := path.Join(loopDevPath, "tpgt_1")
	lunPath := overlaybdLoopbackDeviceLunPath(loopDevID)
	err = os.MkdirAll(lunPath, 0700)
	if err != nil {
		return "", fmt.Errorf("failed to create loopback lun dir %s: %w", lunPath, err)
	}

	defer func() {
		if retErr != nil {
			rerr := os.RemoveAll(lunPath)
			if rerr != nil {
				log.G(ctx).Warnf("failed to clean loopback lun %s, err %v", lunPath, rerr)
			}

			rerr = os.RemoveAll(tpgtPath)
			if rerr != nil {
				log.G(ctx).Warnf("failed to clean loopback tpgt %s, err %v", tpgtPath, rerr)
			}

			rerr = os.RemoveAll(loopDevPath)
			if rerr != nil {
				log.G(ctx).Warnf("failed to clean loopback dir %s, err %v", loopDevPath, rerr)
			}
		}
	}()

	nexusPath := path.Join(tpgtPath, "nexus")
	err = ioutil.WriteFile(nexusPath, ([]byte)(loopDevID), 0666)
	if err != nil {
		return "", fmt.Errorf("failed to write loopback nexus %s: %w", nexusPath, err)
	}

	linkPath := path.Join(lunPath, "dev_"+snID)
	err = os.Symlink(targetPath, linkPath)
	if err != nil {
		return "", fmt.Errorf("failed to create loopback link %s: %w", linkPath, err)
	}

	elapsed = float64(time.Since(starttime)) / float64(time.Millisecond)
	log.G(ctx).Debugf("to tcm loop started time used: %v", elapsed)

	defer func() {
		if retErr != nil {
			rerr := os.RemoveAll(linkPath)
			if err != nil {
				log.G(ctx).Warnf("failed to clean loopback link %s, err %v", linkPath, rerr)
			}
		}
	}()

	devAddressPath := path.Join(tpgtPath, "address")
	bytes, err := ioutil.ReadFile(devAddressPath)
	if err != nil {
		return "", fmt.Errorf("failed to read loopback address for %s: %w", devAddressPath, err)
	}
	deviceNumber := strings.TrimSuffix(string(bytes), "\n")
	log.G(ctx).Infof("get device number %s", deviceNumber)

	// The device doesn't show up instantly. Need retry here.
	var lastErr error = nil
	for retry := 0; retry < maxAttachAttempts; retry++ {
		devDirs, err := ioutil.ReadDir(scsiBlockDevicePath(deviceNumber))
		if err != nil {
			lastErr = err
			time.Sleep(5 * time.Millisecond)
			continue
		}
		if len(devDirs) == 0 {
			lastErr = errors.Errorf("empty device found")
			time.Sleep(5 * time.Millisecond)
			continue
		}

		for _, dev := range devDirs {
			device := fmt.Sprintf("/dev/%s", dev.Name())
			return device, nil
		}
	}
	log.G(ctx).Warnf("timeout to find device for snID: %s, lastErr: %v", snID, lastErr)
	return "", lastErr
}

// PrepareMount for overlaybd will prepare a device from provided parents, and return corresponding mount, upper is ignored
func (o *Overlaybd) PrepareMount(ctx context.Context, keyDir string, parents []string, opts ...roDriver.Opt) ([]mount.Mount, error) {
	if _, err := os.Stat(keyDir); err != nil {
		return nil, fmt.Errorf("error stating keyDir: %w", err)
	}

	return []mount.Mount{
		{
			Type:    "bind",
			Source:  keyDir,
			Options: []string{"rbind", "rw"},
		},
	}, nil
}

// Commit for overlaybd will commit the device
func (o *Overlaybd) Commit(ctx context.Context, keyDir string) error {
	tarFile := filepath.Join(keyDir, archive.TarFileName)
	tmpTarFile := filepath.Join(keyDir, tmpTarIndex)
	if _, err := os.Stat(tarFile); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("error stating tar file: %w", err)
	}
	log.G(ctx).Debugf("convert: converting %v %v with args : %v", keyDir, tarFile, tmpTarFile)
	if err := exec.Command(ConverterBinary, tarFile, tmpTarFile, "--export").Run(); err != nil {
		return fmt.Errorf("failed to convert tar file to overlaybd device: %w", err)
	}
	return nil
}

// Cleanup for overlaybd will remove the device
func (o *Overlaybd) Cleanup(ctx context.Context, id string) error {
	loopDevID := overlaybdLoopbackDeviceID(id)
	lunPath := overlaybdLoopbackDeviceLunPath(loopDevID)
	linkPath := path.Join(lunPath, "dev_"+id)
	err := os.RemoveAll(linkPath)
	if err != nil {
		return fmt.Errorf("failed to remove loopback link %s: %v", linkPath, err)
	}
	err = os.RemoveAll(lunPath)
	if err != nil {
		return fmt.Errorf("failed to remove loopback lun %s: %v", lunPath, err)
	}

	loopDevPath := overlaybdLoopbackDevicePath(loopDevID)
	tpgtPath := path.Join(loopDevPath, "tpgt_1")

	err = os.RemoveAll(tpgtPath)
	if err != nil {
		return fmt.Errorf("failed to remove loopback tgpt %s: %w", tpgtPath, err)
	}

	err = os.RemoveAll(loopDevPath)
	if err != nil {
		return fmt.Errorf("failed to remove loopback dir %s: %w", loopDevPath, err)
	}

	err = killProcess(ctx, id, int(syscall.SIGINT))
	if err != nil {
		return fmt.Errorf("failed to kill overlaybd process, sn: %s: %w", id, err)
	}

	targetPath := overlaybdTargetPath(id)
	err = os.RemoveAll(targetPath)
	if err != nil {
		return fmt.Errorf("failed to remove target dir %s: %w", targetPath, err)
	}
	return nil
}

// GetMount for overlaybd will return corresponding mount
func (o *Overlaybd) GetMount(ctx context.Context, keyDir string) ([]mount.Mount, error) {
	devPath := filepath.Join(keyDir, lowerFile)
	if _, err := os.Stat(devPath); err == nil {
		bytes, rerr := ioutil.ReadFile(devPath)
		if rerr != nil {
			return nil, fmt.Errorf("failed to read device file %s: %w", devPath, rerr)
		}
		return []mount.Mount{
			{
				Source:  string(bytes),
				Type:    "ext4",
				Options: []string{"ro"},
			},
		}, nil
	} else if os.IsNotExist(err) {
		return []mount.Mount{
			{
				Source:  keyDir,
				Type:    "bind",
				Options: []string{"rw", "rbind"},
			},
		}, nil
	} else {
		return nil, fmt.Errorf("failed to stat device file %s: %w", devPath, err)
	}
}

// TODO(chaofeng): this function needs optimization
func killProcess(ctx context.Context, id string, signal int) error {
	cmdLine := "ps -ef | grep -v grep | grep -w 'overlaybd-service " + id + "'  | awk '{print $2}' "
	cmd := exec.Command("/bin/bash", "-c", cmdLine)
	out, err := cmd.CombinedOutput()
	strVal := strings.Trim(string(out), "\n ")
	log.G(ctx).Infof("overlaybd: exec ps id:%s, err:%v, out:%s", id, err, strVal)
	if strVal == "" {
		log.G(ctx).Warnf("overlaybd: overlaybd process not found, id: %s", id)
		return nil
	}

	cmdLine = "ps -ef | grep -v grep | grep -w 'overlaybd-service " + id + "'  | awk '{print $2}' | xargs kill -s " + fmt.Sprintf("%d", signal)
	cmd = exec.Command("/bin/bash", "-c", cmdLine)
	out, err = cmd.CombinedOutput()
	log.G(ctx).Infof("overlaybd: exec kill(signal: %d) id: %s, pid:%s,  err:%v, out:%s", signal, id, strVal, err, string(out))
	if signal == int(syscall.SIGUSR2) {
		return nil
	}
	for i := 0; i < 500; i++ {
		cmdLine = "ps -ef | grep -v grep | grep -w 'overlaybd-service " + id + "' |  wc -l"
		cmd = exec.Command("/bin/bash", "-c", cmdLine)
		out, _ = cmd.CombinedOutput()
		strVal = strings.Trim(string(out), "\n ")
		num, err := strconv.Atoi(strVal)
		if err != nil {
			time.Sleep(20 * time.Microsecond)
			continue
		}
		if num == 0 {
			return nil
		}
		time.Sleep(20 * time.Microsecond)
	}
	log.G(ctx).Warnf("timeout to kill process for id: %s", id)
	return fmt.Errorf("timeout killing overlaybd process for snID: %s", id)
}
