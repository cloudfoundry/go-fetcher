package util

import (
	_ "embed"
	"fmt"
	"html/template"
	"os"
)

const defaultAppName = "code"
const defaultDomain = "cloudfoundry.org"
const defaultMemory = "128M"
const defaultDiskQuota = "128M"
const defaultInstanceCount = "2"

//go:embed config.json.template
var configTemplate string

//go:embed manifest.yml.template
var manifestTemplate string

func GenerateManifest(targetPath string) error {
	t, err := template.New("manifest").Parse(manifestTemplate)
	if err != nil {
		return err
	}

	appName := getEnvOrDefault("APP_NAME", defaultAppName)
	domain := getEnvOrDefault("DOMAIN", defaultDomain)
	route := fmt.Sprintf("%s.%s", appName, domain)

	instances := getEnvOrDefault("INSTANCES", defaultInstanceCount)

	memory := getEnvOrDefault("MEMORY", defaultMemory)

	diskQuota := getEnvOrDefault("DISK_QUOTA", defaultDiskQuota)

	return generateActual(t, targetPath, map[string]any{"appName": appName, "route": route, "memory": memory, "instances": instances, "disk_quota": diskQuota})
}

func GenerateConfig(targetPath string) error {
	t, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return err
	}
	appName := getEnvOrDefault("APP_NAME", defaultAppName)
	domain := getEnvOrDefault("DOMAIN", defaultDomain)
	appDomainName := fmt.Sprintf("%s.%s", appName, domain)

	githubAPIKey := os.Getenv("GITHUB_APIKEY")

	return generateActual(t, targetPath, map[string]any{"appDomainName": appDomainName, "githubAPIKey": githubAPIKey})
}

func generateActual(template *template.Template, templatePath string, mappers any) error {
	f, err := os.Create(templatePath)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck

	err = template.Execute(f, mappers)
	if err != nil {
		return err
	}
	return nil
}

func getEnvOrDefault(envVar string, defaultValue string) string {
	value := os.Getenv(envVar)
	if value == "" {
		value = defaultValue
	}
	return value
}
