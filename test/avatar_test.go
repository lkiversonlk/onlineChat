package test

import (
	"testing"
	"github.com/lkiversonlk/OnlineChat/models"
	"path"
	"io/ioutil"
	"os"
)

func TestAuthAvatar(t *testing.T){
	var authAvatar models.AuthAvatar
	client := new(models.Client)
	url, err := authAvatar.GetAvatarURL(client)

	if err != models.ErrNoAvatarURL {
		t.Error("AuthAvatar.GetAvatarURL should return ErrNoAvatarURL when no value present")
	}

	// set a value
	testUrl := "http://url-to-gravatar/"

	client.UserData = map[string]interface{}{"picture": testUrl}
	url, err = authAvatar.GetAvatarURL(client)
	if err != nil {
		t.Error("AuthAvatar.GetAvatarURL should return no error when value present")
	} else {
		if url != testUrl {
			t.Error("AuthAvatar.GetAvatarURL should return correct URL")
		}
	}
}

func TestGravatarAvatar(t *testing.T) {
	var gravatarAvatar models.GravatarAvatar
	client := new(models.Client)
	client.UserData = map[string]interface{} {
		"email" : "MyEmailAddress@example.com",
	}
	url, err := gravatarAvatar.GetAvatarURL(client)
	if err != nil {
		t.Error("GravatarAvatar.GetAvatarURL should not return a error")
	}

	if url != "//www.gravatar.com/avatar/0bc83cb571cd1c50ba6f3e8a78ef1346" {
		t.Errorf("GravatarAvatar.GetAvatarURL wrongly returned %s", url)
	}
}

func TestFileSystemAvatar(t *testing.T) {
	filename := path.Join("avatars", "abc.jpg")
	ioutil.WriteFile(filename,  []byte{}, 0777)
	defer func(){os.Remove(filename)}()

	var filesystemAvatar FileSystemAvatar
	client := new(models.Client)
	client.UserData = map[string]interface{}{
		"userid" : "abc",
	}

	url, err := filesystemAvatar.GetAvatarURL(client)

	if err != nil {
		t.Error("FileSystemAvatar.GetAvatarURL should not return an error")
	}

	if url != "/avatars/abc.jpg" {
		t.Errorf("FileSystemAvatar.GetAvatarURL wrongly returned %s", url)
	}
}