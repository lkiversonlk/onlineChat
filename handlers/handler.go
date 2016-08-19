package handlers

import (
	"net/http"
	"io"
	"io/ioutil"
	"path"
)

type ChainHandler struct {
	next http.Handler
}

func NewChainedHandler(handler http.Handler) *ChainHandler{
	return &ChainHandler{handler}
}

func (c *ChainHandler) Next() http.Handler{
	return c.next
}

func UploaderHandle(w http.ResponseWriter, r *http.Request) {
	userId := r.FormValue("userid")
	file, header, err := r.FormFile("avatarFile")

	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	data, err := ioutil.ReadAll(file)

	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	filename := path.Join("avatars", userId + path.Ext(header.Filename))
	err = ioutil.WriteFile(filename, data, 0777)

	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	w.Header()["Location"] = []string{"/"}
	w.WriteHeader(http.StatusTemporaryRedirect)
}