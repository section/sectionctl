package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type PackageJSON struct {
	Name         string            `json:"name"`
	Private      bool              `json:"private"`
	Version      string            `json:"version"`
	TCPPort      string               `json:"tcp_port"`
	Dependencies map[string]string `json:"dependencies"`
	Section      struct {
		AccountID   string `json:"accountId"`
		AppID       string `json:"appId"`
		Environment string `json:"environment"`
		StartScript string `json:"start-script"`
	} `json:"section"`
	X map[string]interface{} `json:"-"` // Rest of the fields should go here.
}
type SectionConfigJSON struct {
	Proxychain []struct {
		Name  string `json:"name"`
		Image string `json:"image"`
	} `json:"proxychain"`
	Environments struct {
		Production struct {
			Origin struct {
				Address string `json:"address"`
			} `json:"origin"`
			IPBlacklist []string `json:"ip_blacklist"`
		} `json:"Production"`
		Development struct {
			Origin struct {
				Address string `json:"address"`
			} `json:"origin"`
		} `json:"Development"`
	} `json:"environments"`
	X map[string]interface{} `json:"-"` // Rest of the fields should go here.
}

func ParsePackageJSON(packageJSONContents string) (PackageJSON, error) {
	packageJSONContent := new(bytes.Buffer)
	if err := json.Compact(packageJSONContent, []byte(packageJSONContents)); err != nil {
		return PackageJSON{}, err
	}
	packageJSON := PackageJSON{}
	dec := json.NewDecoder(strings.NewReader(string(packageJSONContent.Bytes())))
	if err := dec.Decode(&packageJSON); err != nil {
		return PackageJSON{}, fmt.Errorf("Failed to decode JSON: %w", err)
	}
	return packageJSON, nil
}
func ParseSectionConfig(sectionConfigContents string) (SectionConfigJSON, error) {
	sectionConfigContent := new(bytes.Buffer)
	if err := json.Compact(sectionConfigContent, []byte(sectionConfigContents)); err != nil {
		fmt.Println(err)
	}
	sectionConfig := SectionConfigJSON{}
	dec := json.NewDecoder(strings.NewReader(string(sectionConfigContent.Bytes())))
	if err := dec.Decode(&sectionConfig); err != nil {
		return SectionConfigJSON{}, fmt.Errorf("Failed to decode JSON: %v", err)
	}
	return sectionConfig, nil
}
