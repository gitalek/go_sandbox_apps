package main

import (
	"flag"
	"github/gitalek/go_sandbox_apps/webchat/pkg/types"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
)

//const host = "localhost"
//const port = 9090
const templDir = "webchat/templates"

type templateHandler struct {
	once sync.Once
	templDir string
	filename string
	templ *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(t.pathToFile()))
	})
	err := t.templ.Execute(w, r)
	if err != nil {
		log.Printf("Error occured while executing template: %v", err)
	}
}

func (t *templateHandler) pathToFile() string {
	return filepath.Join(t.templDir, t.filename)
}

func main() {
	addr := flag.String("addr", ":9090", "The addr of the application")
	flag.Parse()
	//addr := fmt.Sprintf("%s:%d", host, port)
	r := types.NewRoom()
	//r.Tracer = trace.New(os.Stdout)

	// root
	http.Handle("/", &templateHandler{templDir: templDir, filename: "chat.html"})
	// joining the room
	http.Handle("/room", r)

	go r.Run()

	log.Printf("starting server on %s\n", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("ListenAndServe: %v\n", err)
	}
}
