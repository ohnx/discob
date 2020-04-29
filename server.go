package main

import (
    // standard library
    "net/http"
    "log"
    "strconv"
    "fmt"
    "strings"
    "path/filepath"
    "flag"
    "html/template"
)

var githelper GitHelper
var lsttmpl *template.Template

type TemplateHelper struct {
    Path string
    Files []GitFile
}

func GuessMimeType(filename string, file string) string {
    switch filepath.Ext(filename) {
    case ".css": return "text/css"
    case ".coffee": return "text/coffeescript"
    case ".eot": return "application/vnd.ms-fontobject"
    case ".htm": return "text/html"
    case ".html": return "text/html"
    case ".ics": return "text/calendar"
    case ".js": return "application/javascript"
    case ".json": return "application/json"
    case ".markdown": return "text/markdown"
    case ".md": return "text/markdown"
    case ".otf": return "font/otf"
    case ".pdf": return "application/pdf"
    case ".svg": return "image/svg+xml"
    case ".swf": return "application/x-shockwave-flash"
    case ".wasm": return "application/wasm"
    case ".woff": return "application/font-woff"
    case ".woff2": return "application/font-woff2"
    case ".xml": return "text/xml"
    case ".yml": return "text/yaml"
    default: return ""
    }
}

func RequestHandler(w http.ResponseWriter, r *http.Request) {
    domainParts := strings.Split(r.Host, ".")
    if len(domainParts) > 1 {
        revid, err := githelper.FetchRevision(domainParts[0])

        if err != nil {
            w.WriteHeader(http.StatusNotFound)
            w.Header().Add("Content-Type", "text/plain")
            fmt.Fprintf(w, "Commit not found: %s\n", err)
            return
        }

        if r.URL.Path[len(r.URL.Path)-1:] == "/" {
            // ends with forward slash, so this is a directory
            dir := r.URL.Path[:len(r.URL.Path)-1]
            if len(dir) > 0 {
                dir = dir[1:]
            }
            result, err := githelper.FetchTreeAtRevision(revid, dir)
            if err != nil {
                w.WriteHeader(http.StatusNotFound)
                w.Header().Add("Content-Type", "text/plain")
                fmt.Fprintf(w, "Directory not found: %s\n", err)
                return
            }

            w.Header().Add("Content-Type", "text/html")
            err = lsttmpl.Execute(w, TemplateHelper{
                Path: r.URL.Path,
                Files: result,
            })

            if err != nil {
                fmt.Fprintf(w, "Error executing template: %s\n", err)
            }
        } else {
            result, err := githelper.FetchFileAtRevision(revid, r.URL.Path[1:])
            if err != nil {
                w.WriteHeader(http.StatusNotFound)
                w.Header().Add("Content-Type", "text/plain")
                fmt.Fprintf(w, "File not found: %s\n", err)
                return
            }

            mimeGuess := GuessMimeType(r.URL.Path, result)
            if len(mimeGuess) > 0 {
                // only set it if the guess was successful, otherwise let
                // the internal server deal with it
                w.Header().Set("Content-Type", mimeGuess)
            }
            fmt.Fprint(w, result)
        }
    } else {
        w.WriteHeader(http.StatusNotFound)
        w.Header().Add("Content-Type", "text/plain")
        fmt.Fprintf(w, "I have no idea what to serve you\n")
    }
}

func main() {
    // Flags
    port := flag.Int("port", 8080, "Port to listen on")
    host := flag.String("host", "", "Host to bind to")
    repo := flag.String("repo", ".", "Repository to serve")
    tmpl := flag.String("tmpl", "directory.html", "Template HTML to use for directory listings")
    flag.Parse()

    // initialize git stuff
    githelper.InitGitHelper(*repo)

    // set up template
    lsttmpl = template.Must(template.ParseFiles(*tmpl))

    // Start server
    http.HandleFunc("/", RequestHandler)
    log.Printf("Server listening on %s:%d", *host, *port)
    err := http.ListenAndServe(*host + ":" + strconv.Itoa(*port), nil)
    if err != nil {
        log.Fatalf("Failed to listen: %s", err)
    }
}
