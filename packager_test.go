package main

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

func TestPackage(t *testing.T) {
	customScript := NewCustomScript()
	customScript.SetDiscordTag("User#1234")
	reader, err := zip.OpenReader("HuokanAdvertiserTools.zip")
	if err != nil {
		t.Errorf("failed to open zip file: %v", err)
		return
	}

	addon, err := Package(&reader.Reader, customScript)
	if err != nil {
		t.Errorf("Failed to package addon: %v", err)
		return
	}

	if addon.Version != "v1.0.0" {
		t.Errorf("Expected version v1.0.0, got %v", addon.Version)
	}

	reader2, err := zip.NewReader(bytes.NewReader(addon.Content), int64(len(addon.Content)))
	if err != nil {
		t.Errorf("Failed to read newly created zip: %v", err)
		return
	}
	f, err := reader2.Open("HuokanAdvertiserTools/Custom.lua")
	if err != nil {
		t.Errorf("Failed to open Custom.lua: %v", err)
		return
	}

	content, err := ioutil.ReadAll(f)
	if err != nil {
		t.Errorf("Failed to read Custom.lua: %v", err)
		return
	}

	if !strings.Contains(string(content), "addon.discordTag = \"User#1234\"") {
		t.Errorf("discordTag not found in Custom.lua")
	}
}
