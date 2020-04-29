# discob

Super-simple web-based Git repository browser.

## how it works

Send a GET request to `<commit identifier>.example.com/path/to/file`

Supported commit identifiers include tags, branches, and commit hashes.

If the file given is a directory (ends in `/`), then it will build a simple directory listing.

### sample GET request

```
GET /server.go
Host: master.discob.git.masonx.ca
```

## options

```
  -host string
        Host to bind to
  -port int
        Port to listen on (default 8080)
  -repo string
        Repository to serve (default ".")
  -tmpl string
        Template HTML to use for directory listings (default "directory.html")
```

## libraries used

`go-git`: Git interface

## how to run

Needs Go 1.11 (modules) support.

```
export GO111MODULE="on"
go build github.com/ohnx/discob
```
