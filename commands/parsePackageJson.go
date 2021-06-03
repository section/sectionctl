package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

type PackageJSON struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
	Scripts      map[string]string `json:"scripts"`
	Section      struct {
		AccountID   string `json:"accountId"`
		AppID       string `json:"appId"`
		Environment string `json:"environment"`
		ModuleName  string `json:"module-name"`
		StartScript string `json:"start-script"`
	} `json:"section"`
	X map[string]interface{} `json:"-"` // Rest of the fields should go here.
}
type PotentialIntValues struct {
	Section struct {
		AccountID int `json:"accountId"`
		AppID     int `json:"appId"`
	} `json:"section"`
}
type MinimalPackageJSON struct {
	Section struct {
		AccountID   string `json:"accountId"`
		AppID       string `json:"appId"`
		Environment string `json:"environment"`
		ModuleName  string `json:"module-name"`
		StartScript string `json:"start-script"`
	} `json:"section"`
	X map[string]interface{} `json:"-"` // Rest of the fields should go here.
}
type SectionConfigJSON struct {
	Proxychain []struct {
		Name  string `json:"name"`
		Image string `json:"image"`
	} `json:"proxychain"`
	X map[string]interface{} `json:"-"` // Rest of the fields should go here.
}

// Try to fit the contents of the package.json into one of the three structs defined above, as JSON isn't strictly typed.
func ParsePackageJSON(packageJSONContents string) (PackageJSON, error) {
	packageJSONContent := new(bytes.Buffer)
	if err := json.Compact(packageJSONContent, []byte(packageJSONContents)); err != nil {
		log.Debug().Err(err).Msg("Error compacting json while parsing your package.json")
		return PackageJSON{}, err
	}
	packageJSON := PackageJSON{}
	if err := json.Unmarshal(packageJSONContent.Bytes(), &packageJSON); err != nil {
		potentialIntValues := PotentialIntValues{}
		if err2 := json.Unmarshal(packageJSONContent.Bytes(), &potentialIntValues); err2 != nil {
				log.Debug().Err(err).Err(err2).Msg("Error unmarshaling your package.json")
		}else{
			if potentialIntValues.Section.AccountID != 0 {
				packageJSON.Section.AccountID = strconv.Itoa(potentialIntValues.Section.AccountID)
			}
			if potentialIntValues.Section.AppID != 0 {
				packageJSON.Section.AppID = strconv.Itoa(potentialIntValues.Section.AppID)
			}
		}
	}
	return packageJSON, nil
}

func ParseSectionConfig(sectionConfigContents string) (SectionConfigJSON, error) {
	sectionConfigContent := new(bytes.Buffer)
	if err := json.Compact(sectionConfigContent, []byte(sectionConfigContents)); err != nil {
		log.Debug().Err(err).Msg("Error compacting json while parsing your section.config.json")
	}
	sectionConfig := SectionConfigJSON{}
	dec := json.NewDecoder(strings.NewReader(sectionConfigContent.String()))
	if err := dec.Decode(&sectionConfig); err != nil {
		return SectionConfigJSON{}, fmt.Errorf("failed to decode JSON: %v", err)
	}
	return sectionConfig, nil
}
