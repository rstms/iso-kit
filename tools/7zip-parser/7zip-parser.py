#!/usr/bin/env python3

import sys
import re
import os
import json

import re

def parse_7z_listing(listing_text):
    """
    Parses a 7z 'Listing archive:' text block and returns a list of
    { date, time, attr, size, compressed_size, name, is_directory } dicts.
    """
    # Make date/time optional and allow missing size/compressed fields.
    line_pattern = re.compile(
        r'^\s*'
        r'(?:'                                  # Optional date/time group
        r'(?P<date>\d{4}-\d{2}-\d{2})\s+'
        r'(?P<time>\d{2}:\d{2}:\d{2})'
        r')?\s+'
        r'(?P<attr>[D\.]{5})'                  # Attr is 5 chars of D or .
        r'(?:\s+(?P<size>\d+))?'               # Optional numeric size
        r'(?:\s+(?P<compressed>\d+))?'         # Optional numeric compressed
        r'\s+(?P<name>.+)$'                    # Remainder is the name
    )

    entries = []
    in_listing = False

    for line in listing_text.splitlines():
        # Start parsing after the first line of dashes
        if not in_listing and line.startswith("----------"):
            in_listing = True
            continue

        # End parsing after the next line of dashes
        if in_listing and line.startswith("----------"):
            break

        if not in_listing:
            continue

        match = line_pattern.match(line)
        if not match:
            continue

        # Default missing fields
        date_str = match.group("date") or ""
        time_str = match.group("time") or ""
        attr = match.group("attr") or ""
        size_str = match.group("size") or "0"
        compressed_str = match.group("compressed") or "0"
        name = match.group("name").strip()

        is_directory = attr.startswith("D")

        entries.append({
            "date": date_str,
            "time": time_str,
            "attr": attr,
            "size": int(size_str),
            "compressed_size": int(compressed_str),
            "name": name,
            "is_directory": is_directory
        })

    return entries

def count_dirs_and_files_on_filesystem(path):
    """
    Recursively counts the number of directories and files in a given directory on the filesystem.

    :param path: Filesystem path to a directory to walk through.
    :return: (num_dirs, num_files)
    """
    num_dirs = 0
    num_files = 0
    for _, dirs, files in os.walk(path):
        num_dirs += len(dirs)
        num_files += len(files)
    return num_dirs, num_files

def main():
    if len(sys.argv) < 2:
        print(f"Usage: {sys.argv[0]} <7z_listing_file>")
        sys.exit(1)

    file_path = sys.argv[1]
    with open(file_path, "r", encoding="utf-8") as f:
        listing_text = f.read()

    parsed_entries = parse_7z_listing(listing_text)

    # Print the JSON listing
    print(json.dumps(parsed_entries, indent=2))

    open("parsed.json", "w").write(json.dumps(parsed_entries, indent=2))

    # Count directories vs. files
    num_dirs = sum(1 for e in parsed_entries if e["is_directory"])
    num_files = len(parsed_entries) - num_dirs

    print(f"\nSummary:")
    print(f"  Directories: {num_dirs}")
    print(f"  Files:       {num_files}")

if __name__ == "__main__":
    main()
#     dirs, files = count_dirs_and_files_on_filesystem("/tmp/ubuntu")
#     print(f"Filesystem:")
#     print(f"  Directories: {dirs}\n  Files: {files}")