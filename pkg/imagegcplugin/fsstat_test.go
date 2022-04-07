package imagegcplugin

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/containerd/containerd/mount"
)

func TestCalculateFreeSizeLogic(t *testing.T) {
	ctx := context.TODO()
	cwdPath, err := os.Getwd()
	if err != nil {
		t.Fatal("failed to get cwdpath")
	}

	gcinstance, err := NewImageGarbageCollect(nil, GcPolicy{
		HighThresholdPercent: 100,
		LowThresholdPercent:  1,
		MinAge:               10 * time.Second,
		Whitelist:            []string{"pause"},
	}, map[string]string{"testSnapshotter":cwdPath}, "testSnapshotter")
	if err != nil {
		t.Fatalf("failed to new gc instance: %v", err)
	}

	plugin := gcinstance.(*imageGCHandler)
	gotStat, err := plugin.imageFSStats(ctx)
	if err != nil {
		t.Fatalf("failed to get imageFSStats: %v", err)
	}

	usagePercent := 100 - int(gotStat.availableBytes*100/gotStat.capacityBytes)

	// when highThresholdPercent is lower usagePercent
	{
		if usagePercent < 100 {
			plugin.policy.HighThresholdPercent = usagePercent + 1
		}
		got, err := plugin.calculateFreeSizes(ctx)
		if err != nil {
			t.Fatalf("unexpected error during calculateFreeSizes: %v", err)
		}
		if got != 0 {
			t.Fatalf("expected no disk pressure but got %v bytes need to free", got)
		}
	}
	// when highThresholdPercent is equal to usagePercent
	{
		plugin.policy.HighThresholdPercent = usagePercent
		expectedBytes := (100-1)*gotStat.capacityBytes/100 - gotStat.availableBytes

		got, err := plugin.calculateFreeSizes(ctx)
		if err != nil {
			t.Fatalf("unexpected error during calculateFreeSizes: %v", err)
		}

		// NOTE: The availableBytes might be called changed by other process
		if got != expectedBytes {
			t.Logf("expected free bytes %v but got %v bytes", expectedBytes, got)
		}
	}
}

func TestImageFSStats(t *testing.T) {
	if _, err := exec.LookPath("df"); err != nil {
		if !os.IsNotExist(err) {
			t.Fatalf("failed to find path for df")
		}
		t.Skip("the environment doesn't provide df tools, skip")
	}

	cwdPath, err := os.Getwd()
	if err != nil {
		t.Fatal("failed to get cwdpath")
	}

	info, err := mount.Lookup(cwdPath)
	if err != nil {
		t.Fatalf("failed to get mount info about path: %v", cwdPath)
	}

	ctx := context.TODO()
	plugin := &imageGCHandler{
		imageFSPath:       cwdPath,
		imageFSMountpoint: info.Mountpoint,
	}
	gotStat, err := plugin.imageFSStats(ctx)
	if err != nil {
		t.Fatalf("failed to get imageFSStats: %v", err)
	}
	expectedStat, err := getFsStatsByDF(ctx, cwdPath)
	if err != nil {
		t.Fatalf("failed to get fsStats by df: %v", err)
	}

	t.Logf("from df stats: %v", expectedStat)
	t.Logf("from fsstats syscall stats: %v", gotStat)

	// NOTE: bytes need to be changed into blocks number.
	// And the availableBytes might be called changed by other process
	var errorRangeBlocks uint64 = 1024 * 1024
	if absDiffUint64(gotStat.availableBytes/1024, expectedStat.availableBytes/1024) > errorRangeBlocks {
		t.Logf("big different value for availableBytes between df(%v) and fsstats(%v) result", expectedStat.availableBytes, gotStat.availableBytes)
	}

	if absDiffUint64(gotStat.capacityBytes/1024, expectedStat.capacityBytes/1024) > errorRangeBlocks {
		t.Fatalf("big different value for capacityBytes between df(%v) and fsstats(%v) result", expectedStat.capacityBytes, gotStat.capacityBytes)
	}

	var errorRangeInodes uint64 = 1024
	if absDiffUint64(gotStat.inodes, expectedStat.inodes) > errorRangeInodes {
		t.Fatalf("big different value for inodes between df(%v) and fsstats(%v) result", expectedStat.inodes, gotStat.inodes)
	}

	if absDiffUint64(gotStat.inodesFree, expectedStat.inodesFree) > errorRangeInodes {
		t.Fatalf("big different value for free inodes between df(%v) and fsstats(%v) result", expectedStat.inodesFree, gotStat.inodesFree)
	}
}

func getFsStatsByDF(ctx context.Context, p string) (*fsStats, error) {
	getDFResult := func(args string) (map[string]string, error) {
		res, err := exec.CommandContext(ctx, "df", args, p).CombinedOutput()
		if err != nil {
			return nil, err
		}
		output := strings.TrimSpace(string(res))

		lines := strings.Split(output, "\n")
		if len(lines) != 2 {
			return nil, fmt.Errorf("df should return header and data:\n %s", output)
		}

		// NOTE: since different linux dist might use different version
		// df, use hash4Idx to store the index for the header instead
		// of using df --output=fieldlist
		hash := map[string]string{}

		// NOTE: ignore last "On" column, because "Mount On" will split
		// into two columns.
		header := strings.Fields(strings.TrimSpace(lines[0]))
		data := strings.Fields(strings.TrimSpace(lines[1]))
		if len(header) > len(data) {
			header = header[:len(data)]
		}
		for idx, key := range header {
			hash[strings.ToLower(key)] = data[idx]
		}
		return hash, nil
	}

	getBlocks, err := getDFResult("-k")
	if err != nil {
		return nil, err
	}

	getInodes, err := getDFResult("-i")
	if err != nil {
		return nil, err
	}

	totalBlocks, ok := getBlocks["1k-blocks"]
	if !ok {
		return nil, fmt.Errorf("df return header without 1k-blocks column")
	}

	availableBlocks, ok := getBlocks["available"]
	if !ok {
		return nil, fmt.Errorf("df return header without available column")
	}

	totalInodes, ok := getInodes["inodes"]
	if !ok {
		return nil, fmt.Errorf("df return header without inodes column")
	}

	availableInodes, ok := getInodes["ifree"]
	if !ok {
		return nil, fmt.Errorf("df return header without ifree column")
	}

	return &fsStats{
		availableBytes: uint64(parseStringIntoInt(availableBlocks)) * 1024,
		capacityBytes:  uint64(parseStringIntoInt(totalBlocks)) * 1024,
		inodesFree:     uint64(parseStringIntoInt(availableInodes)),
		inodes:         uint64(parseStringIntoInt(totalInodes)),
	}, nil
}

func parseStringIntoInt(val string) int {
	got, err := strconv.Atoi(val)
	if err != nil {
		panic(fmt.Errorf("failed to parse value into int"))
	}
	return got
}

func absDiffUint64(x, y uint64) uint64 {
	return uint64(math.Abs(float64(x) - float64(y)))
}
