package util

import (
	"fmt"
	"html/template"
	"os"
)

func GenerateManifest(templatePath, targetPath string) error {
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	appName := os.Getenv("APP_NAME")
	if appName == "" {
		return fmt.Errorf("APP_NAME is missing")
	}
	return generateActual(t, targetPath, appName)
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

	return generateActual(t, targetPath, appName+"."+domain)
}

func generateActual(template *template.Template, templatePath string, value string) error {
	f, err := os.Create(templatePath)
	if err != nil {
		return err
	}
	err = template.Execute(f, value)
	if err != nil {
		return err
	}
	return nil
}
