package main

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"regexp"
)

type PackagedAddon struct {
	Version string
	Content []byte
}

func Package(reader *zip.Reader, customScript *CustomScript) (*PackagedAddon, error) {
	output := new(bytes.Buffer)
	writer := zip.NewWriter(output)
	var version string

	for _, inputFile := range reader.File {
		copyAddonFile(writer, inputFile, customScript)
	}
	toc, err := readTOC(reader)
	if err == nil {
		version = getTOCVersion(string(toc))
	} else {
		log.Printf("error reading toc: %v", err)
	}

	err = writer.Close()

	if err != nil {
		return nil, err
	}

	return &PackagedAddon{
		Version: version,
		Content: output.Bytes(),
	}, nil
}

func copyAddonFile(writer *zip.Writer, inputFile *zip.File, customScript *CustomScript) error {
	f, err := writer.Create(inputFile.Name)
	if err != nil {
		return err
	}

	var content []byte
	if inputFile.Name == "HuokanAdvertiserTools/Custom.lua" {
		content = []byte(customScript.GetCustomScript())
	} else {
		content, err = readFileFromZip(inputFile)
		if err != nil {
			return err
		}
	}
	_, err = f.Write(content)
	return err
}

func readFileFromZip(input *zip.File) ([]byte, error) {
	r, err := input.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func readTOC(reader *zip.Reader) ([]byte, error) {
	f, err := reader.Open("HuokanAdvertiserTools/HuokanAdvertiserTools.toc")
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(f)
	return content, err
}

// $ only matches \n, not \r\n, so we need to manually ensure there is no trailing \r
var tocVersionRegex = regexp.MustCompile("(?mi)^## Version: (.*?)\r?$")

func getTOCVersion(content string) string {
	match := tocVersionRegex.FindStringSubmatch(content)
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}
