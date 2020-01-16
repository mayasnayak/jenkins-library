package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	pkgutil "github.com/GoogleContainerTools/container-diff/pkg/util"
)

func test() {

	filename := "go-piper"
	imageName := "docker.wdf.sap.corp:51116/com.sap.piper/go-piper:0.0.1-20190603152528_d979d360c3fe916edee99343f3772d73b48627fb"
	var cachePath string
	cachePath = "./cache"

	image, err := pkgutil.GetImage(imageName, false, cachePath)
	if err != nil {
		fmt.Printf("Error occured %v", err)
	}

	//Join paths
	imageFilePath := filepath.Join(image.FSPath, filename)

	//Get contents of files
	imageFileContents, err := pkgutil.GetFileContents(imageFilePath)
	if err != nil {
		fmt.Printf("Error occured %v", err)
	}

	fileErr := ioutil.WriteFile("test", []byte(*imageFileContents), 0644)
	if fileErr != nil {
		fmt.Printf("Error occured %v", fileErr)
	}
}
