package backend

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"reflect"

	"github.com/BurntSushi/toml"
)

type AppConfig struct {
	MinecraftServerConfig MinecraftServerConfig
	WebAppConfig          WebAppConfig
}

type WebAppConfig struct {
	Port string
}

type MinecraftServerConfig struct {
	PathToMcServer         string
	MaxAlowedRam           string
	MinAlowedRam           string
	ServerJarName          string
	OthersCommandArguments string
}

var SavedAppConfig AppConfig

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func structToMap(s interface{}) map[string]string {
	result := make(map[string]string)
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		result[field.Name] = value.String()
	}

	return result
}

func DecodeConfig() {
	_, err := toml.DecodeFile("./backend/app_settings.toml", &SavedAppConfig)
	Check(err)
}

func EncodeConfig() {
	file, err := os.Create("./backend/app_settings.toml")
	Check(err)
	defer file.Close()

	err = toml.NewEncoder(file).Encode(&SavedAppConfig)
	Check(err)
}

func ChangeAppSettingsHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	setting := r.FormValue("setting")
	value := r.FormValue("value")
	if setting == "" || value == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "missing propertie or value",
		})
		return
	}

	updated := false

	mcConfigValue := reflect.ValueOf(&SavedAppConfig.MinecraftServerConfig).Elem()
	mcConfigType := mcConfigValue.Type()
	for i := 0; i < mcConfigValue.NumField(); i++ {
		if mcConfigType.Field(i).Name == setting {
			field := mcConfigValue.Field(i)
			if field.CanSet() && field.Kind() == reflect.String {
				field.SetString(value)
				updated = true
				break
			}
		}
	}

	if !updated {
		webConfigValue := reflect.ValueOf(&SavedAppConfig.WebAppConfig).Elem()
		webConfigType := webConfigValue.Type()
		for i := 0; i < webConfigValue.NumField(); i++ {
			if webConfigType.Field(i).Name == setting {
				field := webConfigValue.Field(i)
				if field.CanSet() && field.Kind() == reflect.String {
					field.SetString(value)
					updated = true
					break
				}
			}
		}
	}

	if !updated {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid setting name",
		})
		return
	}

	EncodeConfig()
	http.Redirect(w, r, "/settings/view", http.StatusSeeOther)
}

var appSettingsTemplate = template.Must(template.New("app_settings.html").ParseFiles("./frontend/templates/app_settings.html"))

func AppSettingsTableHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	DecodeConfig()

	data := map[string]interface{}{
		"MinecraftServerConfig": structToMap(SavedAppConfig.MinecraftServerConfig),
		"WebAppConfig":          structToMap(SavedAppConfig.WebAppConfig),
	}

	appSettingsTemplate.ExecuteTemplate(w, "app_settings.html", data)
}
