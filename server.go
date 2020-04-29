package main

import (
    // standard library
    "net/http"
    "log"
    "strconv"
    "fmt"
    "strings"
    "flag"
    "html/template"
)

var githelper GitHelper
var lsttmpl *template.Template

type TemplateHelper struct {
    Path string
    Files []GitFile
}

func GitRequestHandler(w http.ResponseWriter, r *http.Request) {
    
}

func DefaultRequestHandler(w http.ResponseWriter, r *http.Request) {
    
}

func RequestHandler(w http.ResponseWriter, r *http.Request) {
    domainParts := strings.Split(r.Host, ".")
    if len(domainParts) > 1 {
        revid, err := githelper.FetchRevision(domainParts[0])

        if err != nil {
            fmt.Fprintf(w, "Commit not found: %s", err)
            goto done
        }

        if r.URL.Path[len(r.URL.Path)-1:] == "/" {
            // ends with forward slash, so this is a directory
            result, err := githelper.FetchTreeAtRevision(revid, r.URL.Path[1:len(r.URL.Path)-1])
            if err != nil {
                fmt.Fprintf(w, "File not found: %s", err)
                goto done
            }

            err = lsttmpl.Execute(w, TemplateHelper{
                Path: r.URL.Path,
                Files: result,
            })

            if err != nil {
                fmt.Fprintf(w, "Error executing template: %s", err)
            }
        } else {
            result, err := githelper.FetchFileAtRevision(revid, r.URL.Path[1:])
            if err != nil {
                fmt.Fprintf(w, "File not found: %s", err)
                goto done
            }
            fmt.Fprintf(w, "%s", result)
        }
    } else {
        fmt.Fprintf(w, "I have no idea what to serve you")
    }

done:
    fmt.Fprintf(w, "\n")
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
    log.Printf("Server listening on 0.0.0.0:%d", *port)
    err := http.ListenAndServe(*host + ":" + strconv.Itoa(*port), nil)
    if err != nil {
        log.Fatalf("Failed to listen: %s", err)
    }
}
