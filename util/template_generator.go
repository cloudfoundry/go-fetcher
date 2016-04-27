package util

import (
	"html/template"
	"os"
  "fmt"
)

func GenerateManifest(filePath string) error {
	t, err := template.ParseFiles(filePath)
	if err != nil {
		return err
	}

  appName := os.Getenv("APP_NAME")
  if appName == "" {
    return fmt.Errorf("APP_NAME is missing")
  }
  return generateActual(t, os.Getenv("ROOT_DIR") + "/manifest.yml", appName)
}

func GenerateConfig(filePath string) error {
	t, err := template.ParseFiles(filePath)
	if err != nil {
		return err
	}
  appName := os.Getenv("APP_NAME")
  domain  := os.Getenv("DOMAIN")
  if appName == "" || domain == "" {
    return fmt.Errorf("APP_NAME or DOMAIN is missing")
  }

  return generateActual(t, os.Getenv("ROOT_DIR") + "/config.json", appName + "." + domain)
}

func generateActual(template *template.Template, filePath string, target string) error {
	f, err := os.Create(filePath)
  if err != nil {
		return err
  }
	err = template.Execute(f, target)
	if err != nil {
		return err
	}
  return nil
}
