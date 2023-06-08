package misc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

type SettingsType struct {
	AssetPath string
}

func GetSettings() SettingsType {
	var settings SettingsType
	err := parseJSONFile(filepath.Join(AppDirectory, "settings.json"), &settings, false)
	if err != nil {
		// fmt.Println("creating settings.json now....")
		settings = SettingsType{
			AssetPath: filepath.Join(AppDirectory, "assets"),
			// AssetPath: filepath.Join(AppDirectory, "./assets"),
		}
		b, jsonMarshalErr := json.Marshal(settings)
		if jsonMarshalErr != nil {
			return SettingsType{} // Return an empty SettingsType on error
		}
		writeErr := os.WriteFile(filepath.Join(AppDirectory, "settings.json"), b, 0644)
		if writeErr != nil {
			return SettingsType{} // Return an empty SettingsType on error
		}
		return settings
	}
	return settings
}

func ChangeSettings(key string, value interface{}) error {
	settings := GetSettings()
	settingsValue := reflect.ValueOf(&settings).Elem()
	if fieldValue := settingsValue.FieldByName(key); fieldValue.IsValid() {
		if fieldValue.Type().AssignableTo(reflect.TypeOf(value)) {
			fieldValue.Set(reflect.ValueOf(value))
		} else {
			return fmt.Errorf("invalid value type for key '%s'", key)
		}
	} else {
		return fmt.Errorf("unknown key '%s'", key)
	}

	b, jsonMarshalErr := json.Marshal(settings)
	if jsonMarshalErr != nil {
		return jsonMarshalErr
	}

	writeErr := os.WriteFile(filepath.Join(AppDirectory, "settings.json"), b, 0644)
	return writeErr
}
