//go:build linux
// +build linux

/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package overlay

//
// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"syscall"
// 	"testing"
//
// 	"github.com/containerd/containerd/mount"
// 	"github.com/containerd/containerd/pkg/testutil"
// 	"github.com/containerd/containerd/snapshots"
// 	"github.com/containerd/containerd/snapshots/overlay/overlayutils"
// 	"github.com/containerd/containerd/snapshots/storage"
// 	"github.com/containerd/containerd/snapshots/testsuite"
// )
//
// func newSnapshotterWithOpts(opts ...Opt) testsuite.SnapshotterFunc {
// 	return func(ctx context.Context, root string) (snapshots.Snapshotter, func() error, error) {
// 		snapshotter, err := NewSnapshotter(root, opts...)
// 		if err != nil {
// 			return nil, nil, err
// 		}
//
// 		return snapshotter, func() error { return snapshotter.Close() }, nil
// 	}
// }
//
// func TestOverlay(t *testing.T) {
// 	testutil.RequiresRoot(t)
// 	t.Run("noOpt", func(t *testing.T) {
// 		newSnapshotter := newSnapshotterWithOpts()
// 		testsuite.SnapshotterSuite(t, "Overlay", newSnapshotter)
// 		t.Run("TestOverlayMounts", func(t *testing.T) {
// 			testOverlayMounts(t, newSnapshotter)
// 		})
// 		t.Run("TestOverlayCommit", func(t *testing.T) {
// 			testOverlayCommit(t, newSnapshotter)
// 		})
// 		t.Run("TestOverlayOverlayMount", func(t *testing.T) {
// 			testOverlayOverlayMount(t, newSnapshotter)
// 		})
// 		t.Run("TestOverlayOverlayRead", func(t *testing.T) {
// 			testOverlayOverlayRead(t, newSnapshotter)
// 		})
// 		t.Run("TestOverlayView", func(t *testing.T) {
// 			testOverlayView(t, newSnapshotter)
// 		})
// 	})
// }
//
// func TestGetActivePath(t *testing.T) {
// 	il := &snapshots.Info{
// 		Name: "testname",
// 	}
// 	il.Labels = make(map[string]string)
// 	if _, err := getActivePath(il, ""); err == nil {
// 		t.Error("should return error when no home in labels")
// 	}
// 	il.Labels[activePath] = "/tmp"
// 	if home, err := getActivePath(il, "testName"); err == nil {
// 		if home != "/tmp/.rwlayer/testName" {
// 			t.Error("home must be /tmp/.rwlayer/testname")
// 		}
// 	} else {
// 		t.Error("should return home when home in labels")
// 	}
// }
//
// func TestUpperPath(t *testing.T) {
// 	il := &snapshots.Info{}
// 	il.Labels = make(map[string]string)
// 	ss := &snapshotter{}
// 	ss.root = "aaa"
// 	id := "xxx"
// 	path := ss.upperPath(nil, id, "testPath")
// 	if path != filepath.Join(ss.root, "snapshots", id, "upper") {
// 		t.Error("bad upper path")
// 	}
// 	path = ss.upperPath(il, id, "testPath")
// 	if path != filepath.Join(ss.root, "snapshots", id, "upper") {
// 		t.Error("bad upper path")
// 	}
// 	home := "/tmp"
// 	il.Labels[activePath] = home
// 	path = ss.upperPath(il, id, "testPath")
// 	if path != filepath.Join(home, ".rwlayer", "testPath", id, "upper") {
// 		t.Errorf("bad upper path")
// 	}
// }
//
// func TestWorkPath(t *testing.T) {
// 	il := &snapshots.Info{}
// 	il.Labels = make(map[string]string)
// 	ss := &snapshotter{}
// 	ss.root = "aaa"
// 	id := "xxx"
// 	path := ss.fsPath(nil, id, "testPath")
// 	if path != filepath.Join(ss.root, "snapshots", id) {
// 		t.Error("bad work path")
// 	}
// 	path = ss.fsPath(il, id, "testPath")
// 	if path != filepath.Join(ss.root, "snapshots", id) {
// 		t.Error("bad work path")
// 	}
// 	home := "/tmp"
// 	il.Labels[activePath] = home
// 	path = ss.fsPath(il, id, "testPath")
// 	if path != filepath.Join(home, ".rwlayer", "testPath", id) {
// 		t.Error("bad work path")
// 	}
// }
//
// func testOverlayMounts(t *testing.T, newSnapshotter testsuite.SnapshotterFunc) {
// 	ctx := context.TODO()
// 	root := t.TempDir()
// 	o, _, err := newSnapshotter(ctx, root)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	mounts, err := o.Prepare(ctx, "/tmp/test", "")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if len(mounts) != 1 {
// 		t.Errorf("should only have 1 mount but received %d", len(mounts))
// 	}
// 	m := mounts[0]
// 	if m.Type != "bind" {
// 		t.Errorf("mount type should be bind but received %q", m.Type)
// 	}
// 	if m.Options[0] != "rbind" {
// 		t.Errorf("expected mount option rw but received %q", m.Options[0])
// 	}
// 	if m.Options[1] != "rw" {
// 		t.Errorf("expected mount option rbind but received %q", m.Options[1])
// 	}
// }
//
// func testOverlayCommit(t *testing.T, newSnapshotter testsuite.SnapshotterFunc) {
// 	ctx := context.TODO()
// 	root := t.TempDir()
// 	o, _, err := newSnapshotter(ctx, root)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	key := "/tmp/test"
// 	mounts, err := o.Prepare(ctx, key, "")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	m := mounts[0]
// 	if err := os.WriteFile(filepath.Join(m.Source, "foo"), []byte("hi"), 0660); err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := o.Commit(ctx, "base", key); err != nil {
// 		t.Fatal(err)
// 	}
// }
//
// func testOverlayOverlayMount(t *testing.T, newSnapshotter testsuite.SnapshotterFunc) {
// 	ctx := context.TODO()
// 	root := t.TempDir()
// 	o, _, err := newSnapshotter(ctx, root)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	key := "/tmp/test"
// 	if _, err = o.Prepare(ctx, key, ""); err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := o.Commit(ctx, "base", key); err != nil {
// 		t.Fatal(err)
// 	}
// 	var mounts []mount.Mount
// 	if mounts, err = o.Prepare(ctx, "/tmp/layer2", "base"); err != nil {
// 		t.Fatal(err)
// 	}
// 	if len(mounts) != 1 {
// 		t.Errorf("should only have 1 mount but received %d", len(mounts))
// 	}
// 	m := mounts[0]
// 	if m.Type != "overlay" {
// 		t.Errorf("mount type should be overlay but received %q", m.Type)
// 	}
// 	if m.Source != "overlay" {
// 		t.Errorf("expected source %q but received %q", "overlay", m.Source)
// 	}
// 	var (
// 		bp    = getBasePath(ctx, o, root, "/tmp/layer2")
// 		work  = "workdir=" + filepath.Join(bp, "work")
// 		upper = "upperdir=" + filepath.Join(bp, "fs")
// 		lower = "lowerdir=" + getParents(ctx, o, root, "/tmp/layer2")[0]
// 	)
//
// 	expected := []string{
// 		"index=off",
// 	}
// 	if !supportsIndex() {
// 		expected = expected[1:]
// 	}
// 	if userxattr, err := overlayutils.NeedsUserXAttr(root); err != nil {
// 		t.Fatal(err)
// 	} else if userxattr {
// 		expected = append(expected, "userxattr")
// 	}
// 	expected = append(expected, []string{
// 		work,
// 		upper,
// 		lower,
// 	}...)
// 	for i, v := range expected {
// 		if m.Options[i] != v {
// 			t.Errorf("expected %q but received %q", v, m.Options[i])
// 		}
// 	}
// }
//
// func getBasePath(ctx context.Context, sn snapshots.Snapshotter, root, key string) string {
// 	o := sn.(*snapshotter)
// 	ctx, t, err := o.ms.TransactionContext(ctx, false)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer t.Rollback()
//
// 	s, err := storage.GetSnapshot(ctx, key)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	return filepath.Join(root, "snapshots", s.ID)
// }
//
// func getParents(ctx context.Context, sn snapshots.Snapshotter, root, key string) []string {
// 	o := sn.(*snapshotter)
// 	ctx, t, err := o.ms.TransactionContext(ctx, false)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer t.Rollback()
// 	s, err := storage.GetSnapshot(ctx, key)
// 	if err != nil {
// 		panic(err)
// 	}
// 	parents := make([]string, len(s.ParentIDs))
// 	for i := range s.ParentIDs {
// 		parents[i] = filepath.Join(root, "snapshots", s.ParentIDs[i], "fs")
// 	}
// 	return parents
// }
//
// func testOverlayOverlayRead(t *testing.T, newSnapshotter testsuite.SnapshotterFunc) {
// 	testutil.RequiresRoot(t)
// 	ctx := context.TODO()
// 	root := t.TempDir()
// 	o, _, err := newSnapshotter(ctx, root)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	key := "/tmp/test"
// 	mounts, err := o.Prepare(ctx, key, "")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	m := mounts[0]
// 	if err := os.WriteFile(filepath.Join(m.Source, "foo"), []byte("hi"), 0660); err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := o.Commit(ctx, "base", key); err != nil {
// 		t.Fatal(err)
// 	}
// 	if mounts, err = o.Prepare(ctx, "/tmp/layer2", "base"); err != nil {
// 		t.Fatal(err)
// 	}
// 	dest := filepath.Join(root, "dest")
// 	if err := os.Mkdir(dest, 0700); err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := mount.All(mounts, dest); err != nil {
// 		t.Fatal(err)
// 	}
// 	defer syscall.Unmount(dest, 0)
// 	data, err := os.ReadFile(filepath.Join(dest, "foo"))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if e := string(data); e != "hi" {
// 		t.Fatalf("expected file contents hi but got %q", e)
// 	}
// }
//
// func testOverlayView(t *testing.T, newSnapshotter testsuite.SnapshotterFunc) {
// 	ctx := context.TODO()
// 	root := t.TempDir()
// 	o, _, err := newSnapshotter(ctx, root)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	key := "/tmp/base"
// 	mounts, err := o.Prepare(ctx, key, "")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	m := mounts[0]
// 	if err := os.WriteFile(filepath.Join(m.Source, "foo"), []byte("hi"), 0660); err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := o.Commit(ctx, "base", key); err != nil {
// 		t.Fatal(err)
// 	}
//
// 	key = "/tmp/top"
// 	_, err = o.Prepare(ctx, key, "base")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := os.WriteFile(filepath.Join(getParents(ctx, o, root, "/tmp/top")[0], "foo"), []byte("hi, again"), 0660); err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := o.Commit(ctx, "top", key); err != nil {
// 		t.Fatal(err)
// 	}
//
// 	mounts, err = o.View(ctx, "/tmp/view1", "base")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if len(mounts) != 1 {
// 		t.Fatalf("should only have 1 mount but received %d", len(mounts))
// 	}
// 	m = mounts[0]
// 	if m.Type != "bind" {
// 		t.Errorf("mount type should be bind but received %q", m.Type)
// 	}
// 	expected := getParents(ctx, o, root, "/tmp/view1")[0]
// 	if m.Source != expected {
// 		t.Errorf("expected source %q but received %q", expected, m.Source)
// 	}
// 	if m.Options[0] != "ro" {
// 		t.Errorf("expected mount option ro but received %q", m.Options[0])
// 	}
// 	if m.Options[1] != "rbind" {
// 		t.Errorf("expected mount option rbind but received %q", m.Options[1])
// 	}
//
// 	mounts, err = o.View(ctx, "/tmp/view2", "top")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if len(mounts) != 1 {
// 		t.Fatalf("should only have 1 mount but received %d", len(mounts))
// 	}
// 	m = mounts[0]
// 	if m.Type != "overlay" {
// 		t.Errorf("mount type should be overlay but received %q", m.Type)
// 	}
// 	if m.Source != "overlay" {
// 		t.Errorf("mount source should be overlay but received %q", m.Source)
// 	}
//
// 	supportsIndex := supportsIndex()
// 	expectedOptions := 2
// 	if !supportsIndex {
// 		expectedOptions--
// 	}
// 	userxattr, err := overlayutils.NeedsUserXAttr(root)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if userxattr {
// 		expectedOptions++
// 	}
//
// 	if len(m.Options) != expectedOptions {
// 		t.Errorf("expected %d additional mount option but got %d", expectedOptions, len(m.Options))
// 	}
// 	lowers := getParents(ctx, o, root, "/tmp/view2")
// 	expected = fmt.Sprintf("lowerdir=%s:%s", lowers[0], lowers[1])
// 	optIdx := 1
// 	if !supportsIndex {
// 		optIdx--
// 	}
// 	if userxattr {
// 		optIdx++
// 	}
// 	if m.Options[optIdx] != expected {
// 		t.Errorf("expected option %q but received %q", expected, m.Options[optIdx])
// 	}
// }
