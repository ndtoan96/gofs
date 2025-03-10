# Gofs

A featureful file server inspired by [dufs](https://github.com/sigoden/dufs).
`gofs` uses very little JavaScript (only for drag and drop), this allows it to
work even on browsers with limited capabilities (for example, browser on some
e-reader devices).

## Demo

https://github.com/user-attachments/assets/901357cb-d227-4369-9d57-e5eec63df241

## Install

### With go

```bash
go install github.com/ndtoan96/gofs@latest
```

### Pre-built binary

You can download pre-built binary in the
[release page](https://github.com/ndtoan96/gofs/releases)

## Features

- [x] Serve static files
- [x] Create folder
- [x] Delete files/folders
- [x] Archive
- [x] Rename
- [x] Upload
- [x] Download
- [x] Copy/Cut/Paste
- [x] Edit
- [x] Support https
- [x] Drag and drop
- [x] Serve index.html
- [x] Sorting
- [x] Searching (plain text search or glob pattern search)
- [x] Preview (markdown, text, code, image, pdf, zip)

## Usage

```
Usage of gofs:
  -d, --dir string        Directory to serve (default ".")
  -h, --host string       Host address to listen (default "[::]")
  -i, --index             Render index.html
  -p, --port int          Port to listen (default 8080)
      --tsl-cert string   Path to an SSL/TLS certificate to serve with HTTPS
      --tsl-key string    Path to an SSL/TLS certificate's private key
  -w, --write             Allow write access
```

## Example

Serve current directory in read-only mode

```bash
gofs
```

Serve current directory in write mode (allow copy, paste, rename, edit,
delete,...)

```bash
gofs -w
```

Serve directory `xxx`

```bash
gofs xxx
```

Use a different port (default is 8080)

```bash
gofs -p 7777
```
