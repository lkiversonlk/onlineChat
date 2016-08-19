package main

import (
	"log"
	"net/http"
	"github.com/lkiversonlk/OnlineChat/models"
	"github.com/lkiversonlk/OnlineChat/trace"
	"flag"
	"os"
	"github.com/lkiversonlk/OnlineChat/handlers"
)



func main() {
	var addr = flag.String("addr", ":80", "The addr of the application.")
	flag.Parse()

	http.Handle("/", handlers.MustAuth(handlers.NewTemplateHandler("chat.html")))
	http.Handle("/login", handlers.NewTemplateHandler("login.html"))
	http.HandleFunc("/auth/", handlers.LoginHandler)
	http.Handle("/upload", handlers.NewTemplateHandler("upload.html"))
	http.HandleFunc("/uploader", handlers.UploaderHandle)
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request){
		http.SetCookie(w, &http.Cookie{
			Name: "auth",
			Value: "",
			Path: "/",
			MaxAge: -1,
		})
		w.Header()["Location"] = []string{"/"}
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	http.Handle("/avatars/", http.StripPrefix("/avatars/", http.FileServer(http.Dir("./avatars"))))

	r := models.NewRoom(models.UseAuthAvatar)
	r.Tracer = trace.New(os.Stdout)
	http.Handle("/room", r)

	go r.Run()

	log.Println("Starting web server on", *addr)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
