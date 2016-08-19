package main

import (
	"log"
	"net/http"
	"github.com/lkiversonlk/OnlineChat/models"
	"github.com/lkiversonlk/OnlineChat/trace"
	"flag"
	"os"
	"github.com/lkiversonlk/OnlineChat/handlers"
	"path/filepath"
	"fmt"
	"encoding/json"
	"io/ioutil"
)

var configFilePath = filepath.Join(".", "conf", "conf.json")

type configuration struct {
	oauth map[string]handlers.AuthConfiguration          `json:"oauth"`
}

func config() *configuration {
	if file, err := os.Open(configFilePath); err != nil {
		panic(fmt.Errorf("fail to open file %s: %s\n", configFilePath, err))
	} else {
		defer file.Close()

		if content, err := ioutil.ReadAll(file); err != nil {
			panic(fmt.Errorf("fail to read configuration file: %s", err))
		} else {
			//fmt.Println(string(content))

			var top map[string] *json.RawMessage
			ret := configuration{
				oauth: map[string]handlers.AuthConfiguration{},
			}

			if err := json.Unmarshal(content, &top); err !=nil {
				panic(fmt.Errorf("fail to parse config file %s: %s\n", configFilePath, err))
			} else {
				for area, message := range(top) {
					switch area {
					case "oauth":
						var oauthMap map[string]*json.RawMessage
						if err := json.Unmarshal(*message, &oauthMap); err !=nil {
							panic(fmt.Errorf("fail to parse config file %s: %s\n", configFilePath, err))
						} else {
							for provider, providerMessage := range(oauthMap) {
								var config handlers.AuthConfiguration
								if err := json.Unmarshal(*providerMessage, &config); err != nil {
									panic(fmt.Errorf("fail to parse config file %s: %s\n", configFilePath, err))
								} else {
									ret.oauth[provider] = config
								}
							}
						}
					}

				}
				return &ret
			}
		}

	}
}

func main() {
	var addr = flag.String("addr", ":3000", "The addr of the application.")
	flag.Parse()

	configuration := config()
	loginHandler := handlers.NewLoginHandler(configuration.oauth)

	http.Handle("/", handlers.MustAuth(handlers.NewTemplateHandler("chat.html")))
	http.Handle("/login", handlers.NewTemplateHandler("login.html"))
	http.Handle("/auth/", loginHandler)
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
