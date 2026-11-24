# FLACMeta

FLACMeta reads and updates common FLAC Vorbis tags and cover art. It retains the
FLAC source reader and copies compressed audio frames unchanged on save. Frame
copying uses a fixed 32 KiB buffer; memory for metadata and artwork is separate
and depends on their size.

## Read, edit, and save

```go
package main

import (
    "os"

    "github.com/gcottom/flacmeta"
)

func copyWithTitle(source, destination, title string) error {
    in, err := os.Open(source)
    if err != nil {
        return err
    }
    defer in.Close()

    tag, err := flacmeta.ReadFLAC(in)
    if err != nil {
        return err
    }
    tag.SetTitle(title)

    // Destination must be a different file. O_EXCL prevents truncating an
    // existing file, including the source, if the wrong path is supplied.
    out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
    if err != nil {
        return err
    }
    saveErr := tag.Save(out)
    closeErr := out.Close()
    if saveErr != nil {
        return saveErr
    }
    return closeErr
}
```

Keep the original seekable input open until all saves finish. Each save starts
from the original frame offset, so sequential saves to different destinations
are supported. Do not use a tag or its input concurrently. Save does not close
either stream. On a write or read error the destination may be incomplete. To
replace an existing file safely, write to a separate temporary file in the same
directory, close it successfully, and then rename it; remove the temporary
file on failure.

Unchanged comment entries (including repeats), the original vendor, and
unrelated metadata blocks are preserved. Changing a supported tag replaces
its original entries with the new value. Set a supported string to empty or
number to zero to clear its entries. When there is a front-cover image,
ReadFLAC selects it over other embedded images. If the selected artwork is
unchanged, its original encoded bytes are preserved. Call SetCoverArt after
editing an image in place to encode it as a new JPEG, or pass nil to remove the
selected embedded image. Other embedded pictures and linked artwork are kept.

The low-level ParseMetadata function reads only metadata. ParseBytes and
FLACFile.Marshal intentionally buffer the whole stream; use ReadFLAC and Save
for large files.

## Acknowledgements

FLACMeta is adapted from [go-flac](https://github.com/go-flac/go-flac).

## License

See [LICENSE.md](LICENSE.md) for the MIT license and the Apache 2.0 notice
covering portions of the project.

## Related libraries

[AudioMeta v3](https://pkg.go.dev/github.com/gcottom/audiometa/v3) ·
[MP3Meta](https://github.com/gcottom/mp3meta) ·
[MP4Meta](https://github.com/gcottom/mp4meta) ·
[OggMeta](https://github.com/gcottom/oggmeta)

## Canonical metadata editor

Use `OpenMetadata` for the shared multi-format field set. It reads ordered
comments, keeps native artwork as encoded bytes, and leaves compressed frames
untouched. Unknown comments, padding, application blocks, cue sheets, seek
tables and reserved metadata blocks remain in their original representation.
`ClearAll(false)` can remove opaque comments and malformed artwork. Application
and reserved blocks are kept because their playback role is unknown.

```go
tag, err := flacmeta.OpenMetadata(in)
if err != nil {
    return err
}
if err := tag.SetField(core.Mood, core.TextValue("Energetic")); err != nil {
    return err
}
return tag.Save(out)
```

Import `github.com/gcottom/audiometa-core` as `core` for the field and value types.
Native PICTURE and comment blocks must fit FLAC's 24-bit length. Duplicate
comment blocks and malformed metadata lengths are rejected. Source and
destination must be distinct files. The source remains caller-owned.
