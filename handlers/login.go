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
	"golang.org/x/net/context"
	"bytes"
	"net/url"
)

type AuthConfiguration struct{
	Id string            `json:"id"`
	Secret string        `json:"secret"`
	RedirectUrl string   `json:"redirectUrl"`
}

type userProfileExtract func(map[string]interface{}) (map[string]interface{}, error)

type oauth2ConfigProxy interface {
	Exchange(context.Context, string) (*oauth2.Token, error)
	AuthCodeURL(string, ...oauth2.AuthCodeOption) string
	//UserInfoRequestUrl(oauth2.Token) string
	AugUrl(string, *oauth2.Token) string
}

type commonOauthProxy struct {
	*oauth2.Config
}

type wechatOauthProxy struct {
	*oauth2.Config
}

func (_ commonOauthProxy)AugUrl(url string, token *oauth2.Token)string {
	return url + "?access_token=" + token.AccessToken
}

func (_ wechatOauthProxy)AugUrl(url string, token * oauth2.Token)string {
	fmt.Println("call wechat AugUrl")
	fmt.Println(token.Extra("access_token"))
	fmt.Println(token.AccessToken)
	if openid, ok := token.Extra("openid").(string); ok {
		fmt.Println("Openid", openid)
		//https://api.weixin.qq.com/sns/userinfo?access_token=ACCESS_TOKEN&openid=OPENID&lang=zh_CN
		return url + "?access_token=" + token.AccessToken + "&openid=" + openid + "&lang=zh_CN"
	} else {
		fmt.Println("extra", token.Extra("openid"))
		return url + "?access_token=" + token.AccessToken + "&lang=zh_CN"
	}

}

func (c wechatOauthProxy) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	var buf bytes.Buffer
	buf.WriteString(c.Endpoint.AuthURL)
	v := url.Values{
		"response_type": {"code"},
		"appid":     {c.ClientID},
		"redirect_uri":  []string{c.RedirectURL},
		"scope":         []string{strings.Join(c.Scopes, " ")},
		"state":         []string{state},
	}
	if strings.Contains(c.Endpoint.AuthURL, "?") {
		buf.WriteByte('&')
	} else {
		buf.WriteByte('?')
	}
	buf.WriteString(v.Encode())
	return buf.String()
}

type loginHandler struct {
	userProfileURL string
	oauthConfig oauth2ConfigProxy
	provider string
	profileExtract userProfileExtract
}

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

func callOauth(oauth oauth2ConfigProxy, code string, url string) (io.ReadCloser, error) {
	if token, error := oauth.Exchange(oauth2.NoContext, code); error != nil {
		log.Fatalf("fail to call %s, error: %s", url, error.Error())
		return nil, ErrOauthAPICall
	} else {
		if response, error := http.Get(oauth.AugUrl(url, token)); error != nil {
			log.Fatalf("fail to call %s, error: %s", url, error.Error())
			return nil, ErrOauthAPICall
		} else {
			return response.Body, nil
		}
	}
}

func (handler *loginHandler) login(w http.ResponseWriter, r *http.Request) {
	url := handler.oauthConfig.AuthCodeURL("random")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (handler *loginHandler) authUser(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if ret, err := callOauth(handler.oauthConfig, code, handler.userProfileURL); err != nil {
	} else {
		defer ret.Close()
		var retDict map[string]interface{}
		if err := json.NewDecoder(ret).Decode(&retDict); err != nil {
			log.Fatalln("Fail to decode user response", "-", err)
		} else {
			if profile, err := handler.profileExtract(retDict); err != nil {
				log.Fatalln("Fail to extract user profile from response", "-", err)
			} else {
				saveAuthAndGoRoot(w, r, profile)
			}
		}
	}
}

func googleLoginHandler(config AuthConfiguration) *loginHandler {

	return &loginHandler{
		userProfileURL: "https://www.googleapis.com/oauth2/v2/userinfo",
		provider: "google",
		oauthConfig: commonOauthProxy{
			&oauth2.Config{
				RedirectURL: config.RedirectUrl,
				ClientID: config.Id,
				ClientSecret: config.Secret,
				Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
				Endpoint:google.Endpoint,
			},
		} ,
		profileExtract: func(result map[string]interface{}) (map[string]interface{}, error) {
			result["userid"] = result["name"]
			return result, nil
		},
	}
}

func bitbucketLoginHandler(config AuthConfiguration) *loginHandler{
	return &loginHandler{
		userProfileURL: "https://api.bitbucket.org/1.0/user",
		provider: "bitbucket",
		oauthConfig: commonOauthProxy{
			&oauth2.Config{
				RedirectURL: config.RedirectUrl,
				ClientID: config.Id,
				ClientSecret: config.Secret,
				Scopes: []string{"account", "email"},
				Endpoint:bitbucket.Endpoint,
			},
		} ,
		profileExtract: func(result map[string]interface{}) (map[string]interface{}, error) {
			if profile, ok := result["user"].(map[string]interface{}); ok {
				profile["userid"] = profile["username"]
				profile["picture"] = profile["avatar"]
				profile["name"] = profile["username"]
				return profile, nil
			} else {
				log.Fatalln("fail to get user segment from bitbucket response")
				return nil, ErrOauthAPICall
			}
		},
	}
}


func wechatLoginHandler(config AuthConfiguration) *loginHandler{
	return &loginHandler{
		userProfileURL: "https://api.weixin.qq.com/sns/userinfo",
		provider: "wechat",
		oauthConfig: wechatOauthProxy{
			&oauth2.Config{
				RedirectURL: config.RedirectUrl,
				ClientID: config.Id,
				ClientSecret: config.Secret,
				Scopes: []string{"snsapi_userinfo"},
				Endpoint:oauth2.Endpoint{
					AuthURL:"https://open.weixin.qq.com/connect/oauth2/authorize",
					TokenURL:"https://api.weixin.qq.com/sns/oauth2/access_token",
				},
			},
		} ,
		profileExtract: func(result map[string]interface{}) (map[string]interface{}, error) {
			result["name"] = result["nickname"]
			result["userid"] = result["nickname"]
			result["picture"] = result["headimgurl"]
			return result, nil
		},
	}
}
type LoginHandler struct{
     handlers map[string](*loginHandler)
}

var ErrUnknownProvider = errors.New("Unknown provider")

func NewLoginHandler(authConfig map[string]AuthConfiguration) (*LoginHandler){
	ret := map[string]*loginHandler{}
	for provider, config := range(authConfig) {
		fmt.Printf("Preare handler for provider %s, id %s, redirect to %s\n", provider, config.Id, config.RedirectUrl)
		switch provider {
		case "google":
			ret["google"] = googleLoginHandler(config)
		case "bitbucket":
			ret["bitbucket"] = bitbucketLoginHandler(config)
		case "wechat":
			ret["wechat"] = wechatLoginHandler(config)
		default:
			fmt.Printf("Unknown configuration for provider %s\n", provider)
		}
	}
	handler := LoginHandler{ret}
	return &handler
}

func (l *LoginHandler) getHandler(provider string) (*loginHandler, error) {

	if handler, ok := l.handlers[provider]; ok {
		return handler, nil
	} else {
		return nil, ErrUnknownProvider
	}
}

func (l *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]

	if handler, err := l.getHandler(provider); err == nil {
		switch action {
		case "login":
			handler.login(w, r)
		case "callback":
			handler.authUser(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "action %s unknown", action)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "provider %s unknown", provider)
	}
}