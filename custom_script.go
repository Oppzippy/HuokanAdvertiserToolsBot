package main

import (
	"fmt"
	"strings"
)

type CustomScript struct {
	customScriptLines []string
}

func NewCustomScript() *CustomScript {
	return &CustomScript{
		customScriptLines: []string{
			"local _, addon = ...\n",
		},
	}
}

func (p *CustomScript) SetDiscordTag(tag string) {
	escaped := strings.Replace(tag, "\"", "\\", -1)
	p.customScriptLines = append(
		p.customScriptLines,
		fmt.Sprintf("addon.discordTag = \"%s\"", escaped),
	)
}

func (p *CustomScript) GetCustomScript() string {
	return strings.Join(p.customScriptLines, "\n")
}
