package internal

import (
	"fmt"
	"os"
	"net/http"
	"io"
)

// download tarball from Github
func DownloadTarball() {

	fileUrl := "https://golangcode.com/images/avatar.jpg"

	err := Downloader("vaultsmith.tar", fileUrl)
	if err != nil {
		fmt.Errorf("error downloading tarball %s", err)
	}


}

func Downloader(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}