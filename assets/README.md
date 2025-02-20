# Assets

Other assets like images, logo and other media files

```bash
xorriso -as mkisofs \
  -volsetid "VOLUMESETID" \
  -publisher "PUBLISHERID" \
  -preparer "DATAPREPID" \
  -appid "APPID " \
  -copyright "COPYFILEID" \
  -abstract "ABSTRACTFILEID" \
  -bibfile "BIBLIOID" \
  -r -J \
  -o output.iso \
  /tmp/test-001
```
