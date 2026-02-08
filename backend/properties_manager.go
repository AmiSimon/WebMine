package backend

import (
	"bufio"
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
			return errors.New("Value should be boolean like type")
		}
		return nil
	}

	if _, err := strconv.Atoi(initialValue); err == nil {
		if _, err := strconv.Atoi(s); err != nil {
			return errors.New("Value should be an integer like type")
		}
		return nil
	}

	if _, err := strconv.ParseFloat(initialValue, 64); err == nil {
		if _, err := strconv.ParseFloat(s, 64); err != nil {
			return errors.New("Value should be a float like type")
		}
		return nil
	}

	return nil
}

func readServerPropertiesFile() map[string]string {
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

	return properties

}

func writeServerPropertiesFile(properties map[string]string, path string) error {
	path = path + "server.properties"
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)

	writer.WriteString("#Minecraft server properties\n")
	writer.WriteString("#" + time.Now().Format(time.UnixDate) + "\n")

	for key, value := range properties {
		line := key + "=" + value + "\n"
		_, err := writer.WriteString(line)
		if err != nil {
			return err
		}
	}

	err = writer.Flush()
	return err
}

func ChangePropertiesHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	property := r.FormValue("property")
	value := r.FormValue("value")

	if property == "" {
		HtmlDetailedError(w, errors.New("Missing property name"))
		return
	}

	err := checkStrType(value, ServerProperties[property])
	if err != nil {
		HtmlDetailedError(w, err)
		return
	}

	ServerProperties[property] = value
	writeServerPropertiesFile(ServerProperties, SavedAppConfig.MinecraftServerConfig.PathToMcServer)
	http.Redirect(w, r, "/properties/view", http.StatusSeeOther)
}

var propertiesTemplate = template.Must(template.New("properties.html").ParseFiles("./frontend/templates/properties.html"))

func PropertiesTableHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	ServerProperties = readServerPropertiesFile()

	propertiesTemplate.ExecuteTemplate(w, "properties.html", ServerProperties)
}
