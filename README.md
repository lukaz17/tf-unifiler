# TFunifiler

T-Force Universal File Operator, cross-platform file management solution.

## Usage

```
unifiler <command> <subcommand> [inputs...] [flags]
```

Run `unifiler --help` to list the commands, and add `--help` to any command to see all its flags. Commands that take files or folders accept them as arguments or with the `-i` flag, and go through folders recursively. In a terminal, TFunifiler shows the progress of long operations.

### Requirements

Some commands run other programs. TFunifiler looks for them on `PATH`, unless you set their location in the settings:

| Program | Used by | Setting |
|---------|---------|---------|
| [7-Zip](https://www.7-zip.org) (`7z`) | `file pack --format 7z` | `paths.x7z` |
| [RAR](https://www.rarlab.com) (`rar`) | `file pack --format rar` | `paths.rar` |
| [FFmpeg](https://ffmpeg.org) (`ffmpeg`) | `video extract-frames` | `paths.ffmpeg` |
| [MediaInfo](https://mediaarea.net/en/MediaInfo) CLI (`mediainfo`) | `video info`, `video extract-frames` | `paths.mediainfo` |

### Configuration

Settings come from the config file `unifiler.yaml`, and environment variables override them; a setting found in neither keeps its default. The config folder holds the config file and the log file, `logs/unifiler.log`, which is rotated automatically:

- Windows: `%APPDATA%\TFunifiler`.
- Linux and macOS: `~/.config/tfunifiler`.
- Portable mode, when a file named `unifiler.portable` sits next to the executable: the folder of the executable.

An environment variable is named `TFUNIFILER_` followed by the key in upper case with `_` in place of `.`, such as `TFUNIFILER_LOG_LEVEL=debug`.

```sh
unifiler config get paths.x7z                   # print the value in use
unifiler config set paths.x7z /opt/7-zip/7zz    # save a value to the config file
unifiler config export                          # print all the settings in use, as YAML
unifiler config export --diff -o settings.yaml  # save only the settings that differ from the defaults
```

`config export` includes `server.apikey` when an API key is set, so keep its output private.

### Workspace

A workspace is a folder where the `metadata` and `mirror` commands keep their data, given with `-w`. It must already exist, and one folder can serve both commands:

- `metadata.db` is the database of `metadata`, created on first use.
- `mirror/` holds the copies made by `mirror`, each named by the SHA-256 of its content, and the rollback files of `mirror scan`, `mirror-<timestamp>.json`.

### Checksum

`checksum create` writes checksum files for files and folders. Each checksum file is written to the `-o` folder, named `[<title>].<algorithm>`, and lists paths relative to that folder as `<hash> *<path>` lines. CRC32 is the exception: it is written as an SFV file, `[<title>].sfv`, with `<path> <hash>` lines.

Supported algorithms: `crc32`, `crc32c`, `md4`, `md5`, `ripemd160`, `sha1` (default), `sha224`, `sha256`, `sha384` and `sha512`.

### File

`file` works on many files at once:

- `hash` prints their hashes.
- `pack` compresses them into 7z or RAR archives.
- `rename` performs multi-rename for files using pre-defined presets. It takes files only: folders given as inputs are ignored. Before renaming, it saves the old and new names to a rollback file, `unifiler-file-rename-<timestamp>.json`, in the current folder.

> [!WARNING]
> `file pack --move` and `file pack --normalize` change your original files: `--move` deletes the inputs once they are packed, and `--normalize`, before packing, sets their dates to 2020-02-02 12:00 UTC and, on Windows, turns on their Archive attribute and turns off their Read-only, Hidden and System attributes.

### Metadata

`metadata` keeps the hashes of known files in a database in a [workspace](#workspace), grouped into named collections:

- `index` records the hashes, size and location of files in the workspace's database, `metadata.db`.
- `refine` moves each file found in the collections into a `.backup` folder next to it, or with `--invert`, each file not found into `.extra`. With `--erase`, it deletes them instead, which cannot be undone.
- `query` shows the collections, their files, hashes and indexing sessions stored in `metadata.db`, so you don't need to install an external database tool to read it.

### Mirror

`mirror` keeps one copy of each distinct file in a [workspace](#workspace) to save space. Files are told apart by the SHA-256 of their content.

- `scan` adds files to the workspace, and saves each file's path and SHA-256 to a rollback file there.
- `export` rebuilds a folder tree from the workspace, following a SHA-256 checksum file such as one made by `checksum create`.

> [!NOTE]
> The copies are hard links, not real copies: the workspace and the files must be on the same drive, and editing a file in place changes it everywhere it is linked, the workspace included.

### Video

`video` works on a single video file:

- `extract-frames` saves JPEG frames taken across the whole video, converting HDR to SDR colors.
- `info` saves its media information to `<file>.mediainfo.json` next to it.

## License

TFunifiler is licensed under GPL-3.0 license. See [COPYING](COPYING) file for details. TFunifiler is developed using the following libraries and frameworks:

- [nullable](https://github.com/tee8z/nullable) library, Copyright 2020 Athaariq Ardhiansyah. Licensed under MIT license.

TFunifiler Build Library is licensed under LGPL-3.0 license. See [COPYING.LESSER](COPYING.LESSER) file for details. Please notice that a portion of TFunifiler Build Library is a modified version of the following libraries:

- [The go-ethereum](https://github.com/ethereum/go-ethereum) library, Copyright 2016 The go-ethereum Authors. Licensed under LGPL-3.0 license.
