package overlaybd

import (
	"fmt"
	"os"

	"github.com/containerd/containerd/snapshots/overlay/roDriver"
)

type Overlaybd struct {
}

func New() (roDriver.RoDriver, error) {
	if err := SupportsOverlaybd(); err != nil {
		return nil, err
	}
	return &Overlaybd{}, nil
}

func SupportsOverlaybd() error {
	// 1. ensure overlaybd-service exist
	_, err := os.Stat(ServiceBinary)
	if err != nil {
		return fmt.Errorf("error stating overlaybd service binary: %w", err)
	}
	// 2. ensure overlaybd converters exist
	_, err = os.Stat(ConverterBinary)
	if err != nil {
		return fmt.Errorf("error stating overlaybd converter binary: %w", err)
	}
	_, err = os.Stat(ConverterMergeBinary)
	if err != nil {
		return fmt.Errorf("error stating overlaybd merger binary: %w", err)
	}
	return nil
}
