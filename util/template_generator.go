package util

import (
	"fmt"
	"html/template"
	"os"
	"strings"
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

        domain := os.Getenv("DOMAIN")
        if domain == "" {
                return fmt.Errorf("DOMAIN is missing")
        }

        route := appName + "." + domain

        services := os.Getenv("SERVICES")
        serviceNames := strings.Split(services, ",")
        if len(services) == 0 {
                serviceNames = nil
        }

        instances := os.Getenv("INSTANCES")
        if instances == "" {
                instances = "1"
        }

        memory := os.Getenv("MEMORY")
        if memory == "" {
                memory = "512M"
        }

        disk_quota := os.Getenv("DISK_QUOTA")
        if disk_quota == "" {
                disk_quota = "512M"
        }

        return generateActual(t, targetPath, map[string]interface{}{"appName": appName, "services": serviceNames, "route": route, "memory": memory, "instances": instances, "disk_quota": disk_quota})
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

	githubAPIKey := os.Getenv("GITHUB_APIKEY")

	return generateActual(t, targetPath, map[string]interface{}{"appDomainName": appName + "." + domain, "githubAPIKey": githubAPIKey})
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
