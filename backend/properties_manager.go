package backend

import (
	"bufio"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var ServerProperties = make(map[string]string)

func checkStrType(s string, initialValue string) error {
	if initialValue == "true" || initialValue == "false" {
		if s != "true" && s != "false" {
			return errors.New("String value should be Bool like type")
		}
		return nil
	}

	if _, err := strconv.Atoi(initialValue); err == nil {
		if _, err := strconv.Atoi(s); err != nil {
			return errors.New("String value should be Int like type")
		}
		return nil
	}

	if _, err := strconv.ParseFloat(initialValue, 64); err == nil {
		if _, err := strconv.ParseFloat(s, 64); err != nil {
			return errors.New("String value should be Float like type")
		}
		return nil
	}

	return nil
}

func readServerPropertiesFile() {
	path := SavedAppConfig.MinecraftServerConfig.PathToMcServer + "server.properties"

	f, err := os.Open(path)
	Check(err)
	defer f.Close()

	properties := make(map[string]string)
	scanner := bufio.NewScanner(f)

	scanner.Scan()
	scanner.Scan()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			properties[parts[0]] = parts[1]
		}
	}

	Check(scanner.Err())

	ServerProperties = properties

}

func writeServerPropertiesFile(properties map[string]string, path string) {
	path = path + "server.properties"
	f, err := os.Create(path)
	Check(err)
	defer f.Close()

	writer := bufio.NewWriter(f)

	writer.WriteString("#Minecraft server properties\n")
	writer.WriteString("#" + time.Now().Format(time.UnixDate) + "\n")

	for key, value := range ServerProperties {
		line := key + "=" + value + "\n"
		_, err := writer.WriteString(line)
		Check(err)
	}

	err = writer.Flush()
	Check(err)
}

func ChangePropertiesHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	property := r.FormValue("property")
	value := r.FormValue("value")
	jsonEncoder := json.NewEncoder(w)

	if property == "" || value == "" {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(map[string]string{
			"error": "missing property or value",
		})
		return
	}

	err := checkStrType(value, ServerProperties[property])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	ServerProperties[property] = value
	writeServerPropertiesFile(ServerProperties, SavedAppConfig.MinecraftServerConfig.PathToMcServer)
	http.Redirect(w, r, "/properties/view", http.StatusSeeOther)
}

var propertiesTemplate = template.Must(template.New("properties.html").ParseFiles("./frontend/templates/properties.html"))

func PropertiesTableHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	readServerPropertiesFile()

	propertiesTemplate.ExecuteTemplate(w, "properties.html", ServerProperties)
}
