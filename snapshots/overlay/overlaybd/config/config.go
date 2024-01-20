package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/containerd/continuity"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/reference"
)

const (
	gzipIndexFile  = "gzip.meta"
	ext4FSMetaFile = "ext4.fs.meta"
	configFile     = "config.v1.json"
	dataFile       = ".data_file"  // top layer data file for lsmd
	idxFile        = ".data_index" // top layer index file for lsmd
	// labelTurboOCIDigest is the index annotation key for image layer digest
	labelTurboOCIDigest = "containerd.io/snapshot/overlaybd/turbo-oci/target-digest"
	// labelTurboOCIMediaType is the index annotation key for image layer media type
	labelTurboOCIMediaType = "containerd.io/snapshot/overlaybd/turbo-oci/target-media-type"
	overlaybdBaseLayerDir  = "/opt/overlaybd/baselayers"
	overlaybdBaseLayer     = "/opt/overlaybd/baselayers/.commit"
)

// Config is the config of overlaybd target.
type Config struct {
	ImageRef          string        `json:"imageRef"`
	RepoBlobURL       string        `json:"repoBlobUrl"`
	Lowers            []LowerConfig `json:"lowers"`
	Upper             UpperConfig   `json:"upper"`
	ResultFile        string        `json:"resultFile"`
	AccelerationLayer bool          `json:"accelerationLayer,omitempty"`
	RecordTracePath   string        `json:"recordTracePath,omitempty"`
	Proxy             string        `json:"proxy"`
}

// LowerConfig is the config for overlaybd lower layer.
type LowerConfig struct {
	File         string `json:"file,omitempty"`
	Digest       string `json:"digest,omitempty"`
	Size         uint64 `json:"size,omitempty"`
	Dir          string `json:"dir,omitempty"`
	RepoBlobURL  string `json:"repoBlobUrl,omitempty"`
	TargetDigest string `json:"targetDigest,omitempty"` // turboOCI only
	TargetFile   string `json:"targetFile,omitempty"`   // turboOCI only
	GzipIndex    string `json:"gzipIndex,omitempty"`    // turboOCI only
}

// UpperConfig is the config for overlaybd upper layer.
type UpperConfig struct {
	Index string `json:"index,omitempty"`
	Data  string `json:"data,omitempty"`
}

func WriteConfig(config *Config, targetPath string) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal %+v configJSON into JSON: %w", config, err)
	}

	if err := continuity.AtomicWriteFile(targetPath, data, 0600); err != nil {
		return fmt.Errorf("failed to commit the overlaybd config on %s: %w", targetPath, err)
	}
	return nil
}

func ReadConfig(targetPath string) (*Config, error) {
	var ret Config
	data, err := ioutil.ReadFile(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", targetPath, err)
	}

	if err := json.Unmarshal(data, &ret); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data : %w", err)
	}
	return &ret, nil
}

// ConstructTurboOCISpec generates the config spec for TurboOCI only
func ConstructTurboOCISpec(ctx context.Context, parent string, labels map[string]string, snDir, parentDir, blobDigest, blobSize string) error {
	configJSON := Config{
		Lowers:     []LowerConfig{},
		ResultFile: OverlaybdInitDebuglogPath(snDir),
	}

	if parent != "" {
		parentConfJSON, err := ReadConfig(filepath.Join(parentDir, configFile))
		if err != nil {
			return err
		}
		configJSON.RepoBlobURL = parentConfJSON.RepoBlobURL
		configJSON.Lowers = parentConfJSON.Lowers
	} else {
		ref, ok := labels["containerd.io/snapshot/cri.image-ref"]
		if !ok {
			return fmt.Errorf("no image-ref label")
		}
		blobPrefixURL, err := constructImageBlobURL(ref)
		if err != nil {
			return fmt.Errorf("failed to construct image blob prefix url for snapshot: %w", err)
		}
		configJSON.ImageRef = ref
		configJSON.RepoBlobURL = blobPrefixURL
	}

	lower := LowerConfig{
		Dir:          snDir,
		File:         filepath.Join(snDir, ext4FSMetaFile),
		TargetDigest: labels[labelTurboOCIDigest],
	}

	if isGzipLayer(labels[labelTurboOCIMediaType]) {
		lower.GzipIndex = filepath.Join(snDir, gzipIndexFile)
	}
	configJSON.Lowers = append(configJSON.Lowers, lower)
	log.G(ctx).Infof("generating turbooci local config with parent %v, pre lower %v", parent, lower)

	configBuffer, _ := json.MarshalIndent(configJSON, "", "  ")
	log.G(ctx).Debugf("turboOCI local generated config: %s", string(configBuffer))
	return WriteConfig(&configJSON, filepath.Join(snDir, configFile))
}

// ConstructOverlayBDSpec generates the config spec for overlaybd target.
func ConstructOverlayBDSpec(ctx context.Context, parent string, labels map[string]string, snDir, parentDir, blobDigest, blobSize string) error {
	configJSON := Config{
		Lowers:     []LowerConfig{},
		ResultFile: OverlaybdInitDebuglogPath(snDir),
	}
	ref, ok := labels["containerd.io/snapshot/cri.image-ref"]
	if !ok {
		return fmt.Errorf("no image-ref label")
	}
	blobPrefixURL, err := constructImageBlobURL(ref)
	if err != nil {
		return fmt.Errorf("failed to construct image blob prefix url for snapshot: %w", err)
	}
	configJSON.RepoBlobURL = blobPrefixURL
	if parent == "" {
		configJSON.Lowers = append(configJSON.Lowers, LowerConfig{
			Dir: overlaybdBaseLayerDir,
		})
	} else {
		parentConfJSON, err := ReadConfig(filepath.Join(parentDir, configFile))
		if err != nil {
			return err
		}
		if blobPrefixURL == "" {
			configJSON.RepoBlobURL = parentConfJSON.RepoBlobURL
		}
		configJSON.Lowers = parentConfJSON.Lowers
	}

	configJSON.RecordTracePath = ""
	configJSON.ImageRef = ref
	size, _ := strconv.Atoi(blobSize)
	layerConfig := LowerConfig{
		Digest: blobDigest,
		Size:   uint64(size),
		Dir:    snDir,
	}
	configJSON.Lowers = append(configJSON.Lowers, layerConfig)

	return WriteConfig(&configJSON, filepath.Join(snDir, configFile))
}

func ConstructOverlayBDWritableSpec(dir, parent string) error {
	configJSON := Config{
		Lowers:     []LowerConfig{},
		ResultFile: OverlaybdInitDebuglogPath(dir),
	}

	parentConfJSON, err := ReadConfig(filepath.Join(parent, configFile))
	if err != nil {
		return err
	}
	configJSON.RepoBlobURL = parentConfJSON.RepoBlobURL
	configJSON.Lowers = parentConfJSON.Lowers
	configJSON.Upper = UpperConfig{
		Index: path.Join(dir, idxFile),
		Data:  path.Join(dir, dataFile),
	}

	return WriteConfig(&configJSON, filepath.Join(dir, configFile))
}

func constructImageBlobURL(ref string) (string, error) {
	refspec, err := reference.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("invalid repo url %s: %w", ref, err)
	}

	host := refspec.Hostname()
	repo := strings.TrimPrefix(refspec.Locator, host+"/")
	return "https://" + filepath.Join(host, "v2", repo) + "/blobs", nil
}

func isGzipLayer(mediaType string) bool {
	return mediaType == ocispec.MediaTypeImageLayerGzip ||
		mediaType == images.MediaTypeDockerSchema2LayerGzip
}

func OverlaybdInitDebuglogPath(dir string) string {
	return filepath.Join(dir, "init-debug.log")
}

func OverlaybdConfPath(dir string) string {
	return filepath.Join(dir, configFile)
}
