package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
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

// JSON returns a Resolver that retrieves values from a JSON source.
//
// Hyphens in flag names are replaced with underscores.
func PackageJSONResolver(r io.Reader) (kong.Resolver, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r)
	if err != nil{
		return nil,nil
	}
	s := buf.String()
	packageJSON, err:= ParsePackageJSON(s)
	if err != nil {
		log.Info().Err(err).Msg("Error parsing package.json")
	}
	var f kong.ResolverFunc = func(context *kong.Context, parent *kong.Path, flag *kong.Flag) (interface{}, error) {
		accountID,err := strconv.Atoi(packageJSON.Section.AccountID)
		if err == nil{
			if accountID > 0 && flag.Name=="account-id" {
				return packageJSON.Section.AccountID, nil
			}
		}
		appID,err := strconv.Atoi(packageJSON.Section.AppID)
		if err == nil{
			if appID > 0 && flag.Name=="app-id" {
				return packageJSON.Section.AppID, nil
			}
		}
		if len(packageJSON.Section.Environment) > 0 && flag.Name=="environment" {
			return packageJSON.Section.Environment, nil
		}
		if len(packageJSON.Section.ModuleName) > 0 && flag.Name=="app-path" {
			return packageJSON.Section.ModuleName, nil
		}
		return nil, nil
	}

	return f, nil
}
