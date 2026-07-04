# cue
Cue is a declarative cli tool for managing local media assets with YAML. It scans your media library, matches items against upstream databases using NFO metadata, and downloads configured assets to the correct folders. Assets are generic; any combination of source, URL, and filename can be defined, so new asset types like theme music, posters, or trailers can be added without changing any code.

## Databases

Databases are YAML files that describe which assets should be imported for each media item.

Each top-level key is an upstream media ID; a TMDB ID found in the item’s NFO metadata. Cue uses that ID to match a local media folder to an entry in the database, then imports the configured assets into the correct folder.

```yaml
# https://github.com/foo/bar/database.yaml

"584":
  title: 2 Fast 2 Furious
  assets:
    - source: youtube
      url: https://www.youtube.com/watch?v=YvHbvnIttec
      filename: theme.mp3
    - source: gdrive
      url: https://drive.google.com/file/d/example/view
      filename: poster.jpg
```

Currently, YouTube (via `yt-dlp`) and Google Drive are the only accepted sources.

## Configuration

Cue is configured with a YAML file. The config tells Cue which media libraries to scan and which databases to load.

```yaml
# config.example.yaml

libraries:
  - path: /media/movies
    type: movies
  - path: /media/tv
    type: tv

databases:
  - https://github.com/foo/bar/database.yaml # I want this community database for most of my media
  - databases/local.yaml # There are some specific overrides I keep on top
```

Multiple database sources can be used together. When the same media item appears in more than one database, later database files take priority. This allows you to replace or customise assets without editing the original database.

### Flags

Cue currently supports the following flags:

```sh
cue -config config.yaml -down 10
```

| Flag      |       Default | Description                             |
| --------- | ------------: | --------------------------------------- |
| `-config` | `config.yaml` | Path to the Cue configuration file.     |
| `-down`   |          `10` | Maximum number of concurrent downloads. |

## Installation

### Build from source

```sh
git clone https://github.com/santiagosayshey/cue.git
cd cue
go build -o cue .
```

Then run Cue with:

```sh
./cue -config config.yaml
```

You will also need any external tools required by the asset sources you use. For example, YouTube assets require `yt-dlp`.

### Docker

```yaml
# compose.yaml

services:
  cue:
    image: ghcr.io/santiagosayshey/cue:latest
    container_name: cue
    restart: "no"
    volumes:
      - ./config.yaml:/config/config.yaml:ro
      - ./databases:/config/databases:ro
      - /path/to/your/media:/media
```

Cue is designed to run to completion, so the container does not need to stay running. You can run it manually, schedule it with cron, or trigger it from another automation tool.
