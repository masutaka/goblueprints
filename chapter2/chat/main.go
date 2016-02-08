package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
)

// templ は 1 つのテンプレートを表します
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP は HTTP リクエストを処理します
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, data)
}

func main() {
	var addr = flag.String("addr", ":18080", "アプリケーションのアドレス")
	flag.Parse() // フラグを解釈します

	// Gomniauth のセットアップ
	gomniauth.SetSecurityKey("RESIDENCE101")
	gomniauth.WithProviders(
		facebook.New("957653387645069", "bd4b9984868d78cb2012b4554a6c61a8", "http://localhost:18080/auth/callback/facebook"),
		github.New("60c22ff289a776f58746", "cd1581d547e80cfc01586bdfd343317a8b9e3f65", "http://localhost:18080/auth/callback/github"),
		google.New("940801949020-9b6d4r5emh1k6ds9mmmiq1esdt669mrs.apps.googleusercontent.com", "jd3H38hVee1Ragcx0UzStHoW", "http://localhost:18080/auth/callback/google"),
	)

	r := newRoom()
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)

	// チャットルームを開始します
	go r.run()

	// Web サーバーを開始します
	log.Println("Web サーバーを開始します。ポート: ", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
