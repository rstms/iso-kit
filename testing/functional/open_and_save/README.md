# open_and_save

## Overview

`open_and_save` is a functional testing tool designed to validate the integrity of the ISO writing process. It opens an existing Ubuntu ISO image, parses its structure, and then saves it back to a new ISO file. The goal of this test is to ensure that the generated ISO is **bit-for-bit identical** to the original, verifying that the library's write operations correctly preserve all ISO9660 structures, including Rock Ridge, Joliet, and El Torito extensions.

## Purpose

While `open_and_save` also exercises the ISO parsing logic, its **primary focus** is to test the writing (save) functionality. It is meant to be executed **after** testing the parsing and extraction logic, ensuring that an ISO can be loaded and saved **without modification**.

## Features

- Reads an existing Ubuntu ISO image
- Saves it to a new ISO file
- Verifies that the output ISO matches the original (MD5 checksum comparison)
- Supports ISO9660 extensions including:
    - **Rock Ridge** (POSIX-style metadata)
    - **Joliet** (Windows long filenames)
    - **El Torito** (Bootable images)

## Usage

```shell
open_and_save <input.iso> <output.iso>
```

## Example

```shell
open_and_save ubuntu-22.04.iso
```