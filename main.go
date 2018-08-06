package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"trace"

	"github.com/stretchr/objx"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/google"
)

type templateHandler struct {
	once     sync.Once
	fileName string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.fileName)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("go-chat-auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, data)
}

func main() {
	var addr = flag.String("addr", "8080", "アプリケーションのポート番号")
	flag.Parse()

	gomniauth.SetSecurityKey("faiudiopauepoiamfl;adp:zojfdapge:wokgdsaj9")
	gomniauth.WithProviders(
		google.New(
			"591332350862-0ntec6bu8km8st8cr51voannkemvqom1.apps.googleusercontent.com",
			"JRmb5aM9qW4NwCMq_Hyyoddh",
			"http://localhost:8080/auth/callback/google"),
	)

	r := newRoom()
	r.trace = trace.New(os.Stdout)

	http.Handle("/", &templateHandler{fileName: "login.html"})
	http.Handle("/login", &templateHandler{fileName: "login.html"})
	http.Handle("/chat", MustAuth(&templateHandler{fileName: "chat.html"}))
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)
	http.Handle("/assets/",
		http.StripPrefix("/assets",
			http.FileServer(http.Dir("/Users/asobism/dev/go/src/go-chat/assets"))))

	log.Println("開始 port:" + *addr)
	go r.run()

	if err := http.ListenAndServe(":"+*addr, nil); err != nil {
		log.Fatal("サーバ開始エラー", err)
	}

}
