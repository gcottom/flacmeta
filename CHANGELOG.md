# Changelog

## v2.0.0

- Move the Go module to `github.com/gcottom/flacmeta/v2` and use AudioMeta Core v2. Existing API names are available under the new import path.

- Add `OpenMetadata` for the shared registry, multiple pictures, ordered comment
  values, arbitrary custom metadata and read-only STREAMINFO duration.
- Keep raw comment ordering and unfamiliar blocks while streaming the original audio.

These are AI-assisted follow-up changes on the v0.0.7 baseline.
