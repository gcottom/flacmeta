# FLACMETA

FLACMeta is an adaptation of go-flac. The purpose is to provide an interface for audiometa v3. FLACMeta modifies go-flac by not reading the entire file to memory, only the metadata is read to memory. When saving the data, FLACMeta copies stream data directly from the reader to the writer. FLACMeta removes all comment blocks and picture blocks from the file and replaces them with updated blocks. FLACMeta offers reading and writing of many popular vorbis tags. 

## Acknowledgements
[go-flac](github.com/go-flac/go-flac) This library gets a lot of its code and its inspiration from the go-flac library. Thank you go-flac. 

## License
This project is licensed under the MIT License. See the LICENSE file for details. 
Parts of this project are licensed under the Apache 2.0 License. See the LICENSE file for details. 

## Related Links
[audiometa v3](https://github.com/gcottom/audiometa/v3)

[mp3meta](https://github.com/gcottom/mp3meta)

[mp4meta](https://github.com/gcottom/mp4meta)

[oggmeta](https://github.com/gcottom/oggmeta)

