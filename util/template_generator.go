package util

import (
	"fmt"
	"html/template"
	"os"
)

type ManifestMapper struct {
	APP_NAME, SERVICE_NAME string
}

type ConfigMapper struct {
	APPDOMAIN_NAME, GITHUB_APIKEY string
}

func GenerateManifest(templatePath, targetPath string) error {
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	appName := os.Getenv("APP_NAME")
	if appName == "" {
		return fmt.Errorf("APP_NAME is missing")
	}

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		return fmt.Errorf("SERVICE_NAME is missing")
	}
	var appNameMapper = ManifestMapper{appName, serviceName}
	return generateActual(t, targetPath, appNameMapper)
}

func GenerateConfig(templatePath, targetPath string) error {
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}
	appName := os.Getenv("APP_NAME")
	domain := os.Getenv("DOMAIN")
	if appName == "" || domain == "" {
		return fmt.Errorf("APP_NAME or DOMAIN is missing")
	}

	githubApiKey := os.Getenv("GITHUB_APIKEY")
	if githubApiKey == "" {
		return fmt.Errorf("GITHUB_APIKEY is missing")
	}

	var configMapper = ConfigMapper{appName + "." + domain, githubApiKey}
	return generateActual(t, targetPath, configMapper)
}

func generateActual(template *template.Template, templatePath string, mappers interface{}) error {
	f, err := os.Create(templatePath)
	if err != nil {
		return err
	}
	err = template.Execute(f, mappers)
	if err != nil {
		return err
	}
	return nil
}
