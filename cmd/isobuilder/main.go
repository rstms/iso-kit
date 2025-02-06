package main

import (
	"fmt"
	"github.com/bgrewell/iso-kit"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"github.com/bgrewell/iso-kit/pkg/options"
	"os"
)

func main() {

	log := logging.NewSimpleLogger(os.Stderr, logging.TRACE, true)

	img, err := iso.Create("/tmp/ubuntu",
		options.WithLogger(log),
	)
	if err != nil {
		panic(fmt.Errorf("failed to create ISO: %w", err))
	}

	err = img.Write("/tmp/validation.iso")
	if err != nil {
		panic(fmt.Errorf("failed to save ISO: %w", err))
	}

}
