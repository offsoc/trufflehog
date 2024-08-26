package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

type PermissionsData struct {
	Permissions []string `yaml:"permissions"`
	PackageName string   `yaml:"package_name"`
}

const templateText = `// Code generated by go generate; DO NOT EDIT.
package {{ .PackageName }}

import "errors"

type Permission int

const (
    Invalid Permission = iota
{{- range $index, $permission := .Permissions }}
    {{ ToCamelCase $permission }} Permission = iota
{{- end }}
)

var (
    PermissionStrings = map[Permission]string{
{{- range $index, $permission := .Permissions }}
        {{ ToCamelCase $permission }}: "{{ $permission }}",
{{- end }}
    }

    StringToPermission = map[string]Permission{
{{- range $index, $permission := .Permissions }}
        "{{ $permission }}": {{ ToCamelCase $permission }},
{{- end }}
    }

    PermissionIDs = map[Permission]int{
{{- range $index, $permission := .Permissions }}
        {{ ToCamelCase $permission }}: {{ inc $index }},
{{- end }}
    }

    IdToPermission = map[int]Permission{
{{- range $index, $permission := .Permissions }}
        {{ inc $index }}: {{ ToCamelCase $permission }},
{{- end }}
    }
)

// ToString converts a Permission enum to its string representation
func (p Permission) ToString() (string, error) {
    if str, ok := PermissionStrings[p]; ok {
        return str, nil
    }
    return "", errors.New("invalid permission")
}

// ToID converts a Permission enum to its ID
func (p Permission) ToID() (int, error) {
    if id, ok := PermissionIDs[p]; ok {
        return id, nil
    }
    return 0, errors.New("invalid permission")
}

// PermissionFromString converts a string representation to its Permission enum
func PermissionFromString(s string) (Permission, error) {
    if p, ok := StringToPermission[s]; ok {
        return p, nil
    }
    return 0, errors.New("invalid permission string")
}

// PermissionFromID converts an ID to its Permission enum
func PermissionFromID(id int) (Permission, error) {
    if p, ok := IdToPermission[id]; ok {
        return p, nil
    }
    return 0, errors.New("invalid permission ID")
}
`

// ToCamelCase converts a string to CamelCase
func ToCamelCase(s string) string {
	parts := strings.Split(s, ":")
	caser := cases.Title(language.English)
	for i := range parts {
		subParts := regexp.MustCompile(`[\_\.\-]+`).Split(parts[i], -1)
		for j := range subParts {
			subParts[j] = caser.String(subParts[j])
		}
		parts[i] = strings.Join(subParts, "")
	}
	return strings.Join(parts, "")
}

func main() {
	// Read the YAML file from first argument
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to open YAML file: %v", err)
	}
	defer file.Close()

	var data PermissionsData
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		log.Fatalf("Failed to decode YAML file: %v", err)
	}
	data.PackageName = os.Args[3]

	// Parse the template
	tmpl, err := template.New("permissions").Funcs(template.FuncMap{
		"ToCamelCase": ToCamelCase,
		"inc":         func(i int) int { return i + 1 },
	}).Parse(templateText)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	// Generate the code
	outputFile, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()

	err = tmpl.Execute(outputFile, data)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	fmt.Println("Permissions code generated successfully.")
}