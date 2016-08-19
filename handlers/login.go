package handlers

import (
	"net/http"
	"strings"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/bitbucket"
	"log"
	"github.com/stretchr/objx"
	"encoding/json"
	"io"
	"errors"
)

var (
	ripplemasterGoogleID = "364215041019-kofeon73lc382qdc0s16gfrqlamdorth.apps.googleusercontent.com"
	ripplemasterGoogleKey = "HewZPDIkySsyzBWav0TsibC2"
	localGoogleID = "364215041019-ijr0dqtjrapdif57satg9451vbn4g91l.apps.googleusercontent.com"
	localGoogleKey = "I7dIAgUA_53_ht5TmLSPbI3D"
        bitbucketID = "bD8RagSJqnHxUKa4FF"
	bitbucketKey = "qtL5UcxYS4HmSZETggBW3SjxeeVdjmU7"
	localBitbucketID = "HBj7hbYc48UXcBtjk6"
	localBitbucketKey = "k4MdrDwWmH5p3WNeQWj4gL9Vk5VsHyLW"
)

var (
	googleOauth = &oauth2.Config{
		RedirectURL: "http://www.ripplemaster.cn/auth/callback/google",
		ClientID: ripplemasterGoogleID,
		ClientSecret: ripplemasterGoogleKey,
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:google.Endpoint,
	}

	bitbucketOauth = &oauth2.Config{
		RedirectURL: "http://www.ripplemaster.cn/auth/callback/bitbucket",
		ClientID: bitbucketID,
		ClientSecret: bitbucketKey,
		Scopes: []string{"account", "email"},
		Endpoint:bitbucket.Endpoint,
	}
)

func saveAuthAndGoRoot(w http.ResponseWriter, r *http.Request, authCookie map[string]interface{}){
	authCookieValue := objx.New(authCookie).MustBase64()

	http.SetCookie(w, &http.Cookie{
		Name: "auth",
		Value: authCookieValue,
		Path: "/",
	})

	w.Header()["Location"] = []string{"/"}
	w.WriteHeader(http.StatusTemporaryRedirect)
}

var ErrOauthAPICall = errors.New("Fail to get token and call oauth API")

func callOauth(oauth *oauth2.Config, code string, url string) (io.ReadCloser, error) {
	if token, error := oauth.Exchange(oauth2.NoContext, code); error != nil {
		log.Fatalf("fail to call %s, error: %s", url, error.Error())
		return nil, ErrOauthAPICall
	} else {
		if response, error := http.Get(url + "?access_token=" + token.AccessToken); error != nil {
			log.Fatalf("fail to call %s, error: %s", url, error.Error())
			return nil, ErrOauthAPICall
		} else {
			return response.Body, nil
		}
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]
	switch action {
	case "login":
		switch provider {
		case "google":
			url := googleOauth.AuthCodeURL("random")
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		case "bitbucket":
			url := bitbucketOauth.AuthCodeURL("random")
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		default:
			w.WriteHeader(http.StatusNotFound)
		}

	case "callback":
		code := r.FormValue("code")
		var oauth *oauth2.Config
		var userProfileAPI string
		var userProfileExtractor func(map[string]interface{}) (map[string]interface{}, error)

		switch provider {
		case "google":
			oauth = googleOauth
			userProfileAPI = "https://www.googleapis.com/oauth2/v2/userinfo"
			userProfileExtractor = func(result map[string]interface{})(map[string]interface{}, error) {
				result["userid"] = result["name"]
				return result, nil
			}
		case "bitbucket":
			oauth = bitbucketOauth
			userProfileAPI = "https://api.bitbucket.org/1.0/user"
			userProfileExtractor = func(result map[string]interface{})(map[string]interface{}, error) {
				if profile, ok := result["user"].(map[string]interface{}); ok {
					profile["userid"] = profile["username"]
					profile["picture"] = profile["avatar"]
					profile["name"] = profile["username"]
					return profile, nil
				} else {
					log.Fatalln("fail to get user segment from bitbucket response")
					return nil, ErrOauthAPICall
				}
			}
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if ret, err := callOauth(oauth, code, userProfileAPI); err != nil {
			log.Fatalf("fail to call oauth api %s\n", userProfileAPI)
		} else {
			defer ret.Close()
			var retDict map[string]interface{}
			if err := json.NewDecoder(ret).Decode(&retDict); err != nil {
				log.Fatalln("Fail to decode user response", "-", err)
			} else {
				if profile, err := userProfileExtractor(retDict); err != nil {
					log.Fatalln("Fail to extract user profile from response", "-", err)
				} else {
					saveAuthAndGoRoot(w, r, profile)
				}
			}
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s not supported", action)
	}
}
