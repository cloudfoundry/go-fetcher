package util

import (
	"fmt"
	"html/template"
	"os"
)

type TemplateMapper struct {
	APP_NAME, SERVICE_NAME string
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
	var appNameMapper = TemplateMapper{appName, serviceName}
	return generateActual(t, targetPath, appNameMapper)
}

func GenerateConfig(templatePath, targetPath string) error {
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}
	appName := os.Getenv("APP_NAME")
	domain := os.Getenv("DOMAIN")
	var appDomainMapper = TemplateMapper{appName + "." + domain, ""}

	if appName == "" || domain == "" {
		return fmt.Errorf("APP_NAME or DOMAIN is missing")
	}

	return generateActual(t, targetPath, appDomainMapper)
}

func generateActual(template *template.Template, templatePath string, mappers TemplateMapper) error {
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
