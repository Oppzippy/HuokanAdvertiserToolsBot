package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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
		errLogger.Printf("error reading toc: %v", err)
	}

	err = writer.Close()

	if err != nil {
		return nil, fmt.Errorf("error closing zip writer: %v", err)
	}

	return &PackagedAddon{
		Version: version,
		Content: output.Bytes(),
	}, nil
}

func copyAddonFile(writer *zip.Writer, inputFile *zip.File, customScript *CustomScript) error {
	f, err := writer.Create(inputFile.Name)
	if err != nil {
		return fmt.Errorf("error creating file in new zip: %v", err)
	}

	var content []byte
	if inputFile.Name == "HuokanAdvertiserTools/Custom.lua" {
		content = []byte(customScript.GetCustomScript())
	} else {
		content, err = readFileFromZip(inputFile)
		if err != nil {
			return fmt.Errorf("error reading file from HuokanAdvertiserTools zip: %v", err)
		}
	}
	_, err = f.Write(content)
	if err != nil {
		return fmt.Errorf("error writing file to new zip: %v", err)
	}
	return nil
}

func readFileFromZip(input *zip.File) ([]byte, error) {
	r, err := input.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening file in zip: %v", err)
	}
	defer r.Close()

	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading file from zip: %v", err)
	}

	return content, nil
}

func readTOC(reader *zip.Reader) ([]byte, error) {
	f, err := reader.Open("HuokanAdvertiserTools/HuokanAdvertiserTools.toc")
	if err != nil {
		return nil, fmt.Errorf("error opening TOC file in zip: %v", err)
	}
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error reading TOC file in zip: %v", err)
	}
	return content, nil
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
