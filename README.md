# iso-kit

[![codecov](https://codecov.io/gh/bgrewell/iso-kit/graph/badge.svg?token=D15C46IECF)](https://codecov.io/gh/bgrewell/iso-kit)

> **Notice:** This project is in an early development phase and may not yet be fully stable or feature complete. As it evolves, you may encounter significant changes to the API, behavior, and overall functionality.

**iso-kit** is a Go library designed to simplify working with ISO 9660 disk images. Whether you're creating, extracting, or inspecting ISO files, iso-kit provides a reliable and feature-rich solution with advanced support for key extensions like Rock Ridge and El Torito.

In addition to being a library for developers, **iso-kit** also includes some command line tools for working with ISO
files that can be installed via the command line using the commands below.

#### isoextract

**isoextract** is a command line tool for extracting files from an ISO image. It can be installed using the following command:

```bash
go install github.com/bgrewell/iso-kit/cmd/isoextract@latest
```

*note: you may need to ensure that `$GOBIN` is in your `$PATH` you can do that by adding `export PATH=$PATH:$(go env GOPATH)/bin`
to your shell profile.*


## Project Goals

The primary goals of **iso-kit** include:

1. **Comprehensive ISO Handling**: 
   - Support for creating, extracting, and modifying ISO 9660 disk images.
   - Advanced parsing and inspection tools for ISO metadata and structure.

2. **Extension Support**:
   - Full compatibility with Rock Ridge extensions and Joliet for enhanced file attributes.
   - Support for El Torito extensions for handling bootable images.

3. **Simplicity and Usability**:
   - An intuitive API designed for developers.
   - Detailed documentation and examples to accelerate integration.

4. **Performance and Reliability**:
   - Efficient handling of large ISO files.
   - Robust error handling and validation for edge cases.

5. **Future-Proof Design**:
   - Modular and extensible architecture to accommodate future enhancements.
   - Potential for additional features like UDF support or hybrid ISO handling.

---

This library is ideal for developers building tools for ISO manipulation, virtual disk creation, or custom filesystem operations.

Stay tuned for updates as we continue to expand functionality and refine the library!

## ISO9660 Details

### Support

 - [x] ISO 9660
 - [x] El Torito
 - [x] Joliet
 - [x] System Use Sharing Protocol (SUSP)
   - [x] Rock Ridge
   - [ ] CE (SUSP 5.1):
   - [ ] PD (SUSP 5.2):
   - [ ] SP (SUSP 5.3):
   - [ ] ST (SUSP 5.4):
   - [ ] ER (SUSP 5.5):
   - [ ] ES (SUSP 5.6):

### Current Limitations

 - **Extract Only** - Currently this library only supports extraction of files and boot images. Support for creating ISOs is coming soon.
 - **Rock Ridge** - While Rock Ridge is supported, some features may not be fully implemented. Please report any issues you encounter.
 - **Joliet** - Joliet is supported, but some edge cases may not be fully implemented. Please report any issues you encounter.
 - **El Torito** - El Torito is supported, but some edge cases may not be fully implemented. Please report any issues you encounter.
 - **Validation** - This library has not been extensively tested and does not currently have any unit or functional tests so again, report any issues you encouter.

## Test Coverage

<img src="https://codecov.io/gh/bgrewell/iso-kit/graphs/sunburst.svg?token=D15C46IECF"/>
