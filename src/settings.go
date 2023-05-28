package misc

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

func GetSettings() Dict {
	var Settings Dict
	dir := filepath.Join(os.Getenv("LocalAppData"), "sillydokkan")
	if !PathExists(dir) {
		//fmt.Println("settings directory not found, creating it now..")
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	err := parseJSONFile(filepath.Join(dir, "settings.json"), &Settings, false)
	if err != nil {
		//fmt.Println("creating settings.json now....")
		Settings = Dict{
			"AssetPath": "./assets",
		}
		b, jsonMarshallErr := json.Marshal(Settings)
		if jsonMarshallErr != nil {
			return nil
		}
		Writeerr := os.WriteFile(filepath.Join(dir, "settings.json"), b, 0644)
		if Writeerr != nil {
			return nil
		}
		return Settings

	}
	return Settings
}

func ChangeSettings(key string, value any) {
	Settings := GetSettings()
	Settings[key] = value
	b, _ := json.Marshal(Settings)
	_ = os.WriteFile("settings.json", b, 0644)
}
