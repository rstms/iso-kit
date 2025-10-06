package main

import (
	"github.com/rstms/iso-kit"
	"github.com/rstms/iso-kit/pkg/option"
	"os"
)

func main() {

	name := "ubuntu-test-iso"
	source := "/tmp/ubuntu-iso"
	dest := "/tmp/created-ubuntu.iso"
	_ = source //TODO: this will be passed in via an 'AddDir' command
	_ = dest

	// Use values from the real ISO to simplify testing
	name = "Ubuntu-Server 24.04.1 LTS amd64"
	preparer := "XORRISO-1.5.4 2021.01.30.150001, LIBISOBURN-1.5.4, LIBISOFS-1.5.4, LIBBURN-1.5.4"

	i, err := iso.Create(name,
		option.WithPreparerID(preparer),
	)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(dest)
	if err != nil {
		panic(err)
	}

	err = i.Save(f)
	if err != nil {
		panic(err)
	}

}
