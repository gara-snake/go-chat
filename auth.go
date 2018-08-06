package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/stretchr/objx"

	"github.com/stretchr/gomniauth"
)

type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("go-chat-auth"); err != http.ErrNoCookie {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		panic(err.Error())
	} else {
		h.next.ServeHTTP(w, r)
	}
}

// MustAuth is 認証処理
func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	segs := strings.Split(r.URL.Path, "/")
	if len(segs) < 2 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "アクションが指定されていません")
		return
	}

	action := segs[2]
	provider := segs[3]
	switch action {
	case "login":
		prov, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln("プロバイダ選択エラー：", provider, "-", err)
		}
		loginURL, err := prov.GetBeginAuthURL(nil, nil)
		if err != nil {
			log.Fatalln("プロバイダ処理：", provider, "-", err)
		}
		w.Header().Set("Location", loginURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	case "callback":
		prov, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln("認証プロバイダ取得失敗：", provider, "-", err)
		}

		creds, err := prov.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))

		if err != nil {
			log.Fatalln("認証の完了に失敗：", provider, "-", err)
		}

		user, err := prov.GetUser(creds)

		if err != nil {
			log.Fatalln("ユーザ情報の取得に失敗：", provider, "-", err)
		}

		authCookieVale := objx.New(map[string]interface{}{
			"name": user.Name(),
		}).MustBase64()
		http.SetCookie(w, &http.Cookie{
			Name:  "go-chat-auth",
			Value: authCookieVale,
			Path:  "/",
		})
		w.Header()["Location"] = []string{"/chat"}
		w.WriteHeader(http.StatusTemporaryRedirect)

	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "アクション[%s]には非対応です", action)
	}

}
