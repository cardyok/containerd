package sparsefile

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/mount"
)

func (q *sparseFileQuota) createImageFile(img string, size int) (err error) {
	tmpFile, err := ioutil.TempFile(filepath.Dir(img), "new-")
	if err != nil {
		return err
	}
	tmpFileName := tmpFile.Name()
	tmpFile.Close()
	defer func() {
		if err != nil {
			if err1 := os.Remove(tmpFileName); err1 != nil {
				log.G(context.TODO()).WithError(err1).Warnf("failed to remove temp file:%s", tmpFileName)
			}
		}
	}()
	if err := os.Truncate(tmpFileName, int64(size)); err != nil {
		return err
	}
	if err := os.Rename(tmpFileName, img); err != nil {
		return err
	}
	return nil
}

func (q *sparseFileQuota) formatImageFile(img string) error {
	if q.fsType != "ext4" {
		return fmt.Errorf("unsupport %s, only support ext4 file system now", q.fsType)
	}
	args := []string{
		img,
		"-F",
		"-E",
		//TODO(Chaofeng): Disable lazy init to mitigate race condition by compromising on performance.
		// inode number hardcoded 100000 in the future?
		"nodiscard,lazy_itable_init=0",
		"-O",
		"^has_journal",
	}
	output, err := exec.Command("mkfs.ext4", args...).CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "failed to write fs:\n%s", string(output))
	}

	return nil
}

func (q *sparseFileQuota) mount(img string, target string, rwFlag string) error {
	m := &mount.Mount{
		Source: img,
		Type:   defaultFsType,
		Options: []string{
			"loop",
			rwFlag,
		},
	}
	return m.Mount(target)
}

func (q *sparseFileQuota) umount(ctx context.Context, target string) error {
	args := []string{
		"-T",
		target,
		"-P",
		"-o",
		"SOURCE",
	}

	output, err := exec.Command("findmnt", args...).CombinedOutput()
	if err != nil {
		log.G(ctx).Warnf("umount: failed to findmnt target %s: %v", target, err)
	}
	if err := mount.Unmount(target, 0); err != nil {
		return err
	}
	loops := strings.Split(string(output), "\"")
	if len(loops) < 2 {
		log.G(ctx).Infof("umount: findmnt result is: %v", output)
		return nil
	}
	loop := loops[1]
	log.G(ctx).Debugf("umount: trying to ensure %s is cleaned on %s", loop, target)
	for i := 0; i < 30; i++ {
		file, err := os.Open(fmt.Sprintf("/sys/block/%s/loop/backing_file", filepath.Base(loop)))
		if err != nil {
			log.G(ctx).Debugf("umount: open file %s error is: %v", target, err)
			return nil
		}
		buf := make([]byte, 300)
		file.Read(buf)
		file.Close()
		if filepath.Dir(string(buf)) != target {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}
