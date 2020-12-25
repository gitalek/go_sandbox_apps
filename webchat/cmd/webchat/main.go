package main

import (
	"flag"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/objx"
	"github/gitalek/go_sandbox_apps/auth/pkg/auth"
	"github/gitalek/go_sandbox_apps/trace/pkg/trace"
	"github/gitalek/go_sandbox_apps/webchat/pkg/types"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

//const host = "localhost"
//const port = 9090
const templDir = "webchat/templates"

type templateHandler struct {
	once     sync.Once
	templDir string
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(t.pathToFile()))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	err := t.templ.Execute(w, data)
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
	ck, err := ioutil.ReadFile("temp/auth_key.txt")
	if err != nil {
		log.Fatalf("Can't get auth key: %v", err)
	}
	gomniauth.SetSecurityKey(string(ck))
	keyGithub, err := ioutil.ReadFile("temp/github_key.txt")
	if err != nil {
		log.Fatalf("Can't get github key: %v", err)
	}
	secretGithub, err := ioutil.ReadFile("temp/github_secret.txt")
	if err != nil {
		log.Fatalf("Can't get github secret: %v", err)
	}
	gomniauth.WithProviders(
		github.New(
			strings.TrimRight(string(keyGithub), "\r\n"),
			strings.TrimRight(string(secretGithub), "\r\n"),
			"http://localhost:9090/auth/callback/github"),
	)

	r := types.NewRoom()
	r.Tracer = trace.New(os.Stdout)

	// joining the room
	http.Handle("/room", r)
	http.Handle("/chat", auth.MustAuth(&templateHandler{templDir: templDir, filename: "chat.html"}))
	http.Handle("/login", &templateHandler{templDir: templDir, filename: "login.html"})
	http.HandleFunc("/auth/", auth.LoginHandler)
	http.HandleFunc("/logout", func  (w http.ResponseWriter, r *http.Request) {
	    http.SetCookie(w, &http.Cookie{
	    	Name: "auth",
	    	Value: "",
	    	Path: "/",
	    	MaxAge: -1,
		})
	    w.Header().Set("Location", "/chat")
	    w.WriteHeader(http.StatusTemporaryRedirect)
	})

	go r.Run()

	log.Printf("starting server on %s\n", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("ListenAndServe: %v\n", err)
	}
}
