package main

import (
	"archive/zip"
	"bytes"
	"io"
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
		f, err := writer.Create(inputFile.Name)
		if err != nil {
			return nil, err
		}

		var content []byte
		if inputFile.Name == "HuokanAdvertiserTools/Custom.lua" {
			content = []byte(customScript.GetCustomScript())
		} else {
			content, err = readFileFromZip(inputFile)
			if err != nil {
				return nil, err
			}
			if inputFile.Name == "HuokanAdvertiserTools/HuokanAdvertiserTools.toc" {
				version = getTOCVersion(string(content))
			}
		}
		_, err = f.Write(content)
		if err != nil {
			return nil, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	return &PackagedAddon{
		Version: version,
		Content: output.Bytes(),
	}, nil
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

// $ only matches \n, not \r\n, so we need to manually ensure there is no trailing \r
var tocVersionRegex = regexp.MustCompile("(?mi)^## Version: (.*?)\r?$")

func getTOCVersion(content string) string {
	match := tocVersionRegex.FindStringSubmatch(content)
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}
