package main

import "github.com/bgrewell/iso-kit"

func main() {

	source := "/tmp/ubuntu-iso"
	dest := "/tmp/created-ubuntu.iso"
	_ = source //TODO: this will be passed in via an 'AddDir' command

	i, err := iso.Create(dest)
	if err != nil {
		panic(err)
	}

	_ = i

}
