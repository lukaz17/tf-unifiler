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

## Server Mode

Server mode runs TFunifiler as a background server that works through a queue of commands. Add `--queue` to a command to send it to the server instead of running it in your terminal: you get your prompt back right away, and the server runs the queued commands one after another.

Only the `checksum`, `file`, `metadata`, `mirror` and `video` commands can be queued with `--queue`.

### Quick Start

```sh
# Terminal 1: start the server and leave it running (optional, see Auto-start).
unifiler server

# Terminal 2: queue commands, with absolute paths (see Things to Know).
unifiler checksum create /data/photos -o /data/checksums --queue
unifiler file hash /data/archive.7z --queue
```

Each queued command prints `Enqueued task <id> (queued)` and returns without showing progress: the server's terminal shows the current item and how many tasks have finished, and the [HTTP API](#http-api) gives each task's status and output. Stop the server with `unifiler server stop`, or with Ctrl+C in its terminal.

### Commands

`server` manages the server:

- `start` starts the server in this terminal; `unifiler server` alone does the same.
- `stop` asks the running server to stop, and waits until it has.
- `generate-api-key` creates a random API key and saves it to the config file; `--force` replaces a key already set.

### Auto-start

When you queue a command and no server is running, TFunifiler starts one in the background and waits up to 15 seconds for it to answer. It starts in the folder of that command, with its environment, and every later queued command runs there too (see [Things to Know](#things-to-know)). The auto-started server has no window: it writes to the log file (see [Configuration](#configuration)) and keeps running until you run `unifiler server stop`.

Auto-start only starts a server on this machine. Turn it off by setting `server.autostart` to `false`.

### Things to Know

- **Queued commands run in the server's context.** Write them as you would in the server's terminal, with its folder, settings and environment. Your terminal only checks their syntax; other errors show up as `failed` tasks.
- **Client and server must be the same build.** Restart the server after updating TFunifiler.
- **One server per config folder.**
- **Tasks live in memory.** They are lost when the server stops, and while it runs it keeps only the latest `server.backlog` finished tasks and the last 64 KiB of each task's output.
- **Stop the server when idle**, with `pendingTasks` at `0` in `GET /status`. Queued tasks are canceled, and running tasks are cut off after 17 seconds, which can leave their work half done.
- **Canceling only skips queued tasks.** A running task runs to the end.

### Server Settings

Server mode adds these settings to the ones in [Configuration](#configuration), and they are changed the same way. Restart the server after a change.

| Key | Default | Meaning |
|-----|---------|---------|
| `server.http.hostname` | `127.0.0.1` | Address the server listens on, and that `--queue` connects to. |
| `server.http.port` | `17017` | Port the server listens on, and that `--queue` connects to. |
| `server.apikey` | empty | API key the server requires; see [Security](#security). |
| `server.autostart` | `true` | Start a server when a command is queued and none is running. |
| `server.concurrency` | `1` | Number of tasks run at the same time. |
| `server.backlog` | `256` | Number of tasks that can wait to run; more are rejected with `task queue is full`. Also the number of finished tasks kept. |

### Security

Queued commands run with the permissions of the user who started the server, and can read, change and delete that user's files.

- **No API key (default):** the server listens only on this machine (`127.0.0.1`) and rejects requests from web pages, but any user or program on this machine can send it commands. On a shared machine, set an API key.
- **API key:** run `unifiler server generate-api-key`, then restart the server. The key is saved to the config file (on Linux and macOS, readable only by you), where `--queue` and `server stop` on the same machine read it. HTTP clients send it as `Authorization: Bearer <key>`.
- **Listening on the network** (for example `server.http.hostname` set to `0.0.0.0`) requires an API key: the server refuses to start without one. Traffic is plain HTTP, so the key, the commands, including values such as `file pack --password`, and their output cross the network unencrypted. Do this only on a trusted network, or reach the server through an SSH tunnel.
- **Remote clients:** set `server.http.hostname` and `server.apikey` on the client machine to the server's address and key. Paths in queued commands are paths on the server machine, and auto-start does not apply.

### HTTP API

The CLI only queues commands; use the HTTP API to see their results and manage them. Add the `Authorization: Bearer <key>` header when an API key is set.

| Request | Purpose |
|---------|---------|
| `GET /status` | Server version, working directory, config folder and task counts. |
| `GET /tasks` | All tasks, oldest first. |
| `GET /tasks/{id}` | One task, with its output and error. |
| `POST /tasks` | Queue a command; JSON body such as `{"args": ["checksum", "create", "/data/photos", "-o", "/data/checksums"]}`, sent with `Content-Type: application/json`. |
| `POST /tasks/{id}/cancel` | Cancel a task. |
| `POST /tasks/{id}/restore` | Put a task canceled while queued back in the queue, in its old place. |
| `DELETE /tasks` | Remove all finished tasks. |
| `POST /shutdown` | Stop the server. |

For example, to see a task (use `curl.exe` in Windows PowerShell):

```sh
curl http://127.0.0.1:17017/tasks/<id>
```

| Task status | Meaning |
|-------------|---------|
| `queued` | Waiting to run. |
| `canceling` | Canceled while queued; skipped when its turn comes, unless restored first. |
| `running` | Running. |
| `stopping` | Cancel requested while running; it still runs to the end, then becomes `succeeded` or `failed`. |
| `succeeded` | Finished without error. |
| `failed` | Finished with an error, given in `error`. |
| `canceled` | Canceled while queued; it never ran. |

## License

TFunifiler is licensed under GPL-3.0 license. See [COPYING](COPYING) file for details. TFunifiler is developed using the following libraries and frameworks:

- [nullable](https://github.com/tee8z/nullable) library, Copyright 2020 Athaariq Ardhiansyah. Licensed under MIT license.
