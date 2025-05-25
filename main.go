package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	//go:embed "templates/readme_template.md"
	readmeContents string

	//go:embed "templates/adr_template.md"
	adrContents string

	//go:embed "templates/help.txt"
	helpContents string

	//go:embed "templates/readme_templates_folder.md"
	templateReadmeContents string
)

func rebuildReadme() error {
	files, err := getAllFilesInAdrDir()
	if err != nil {
		return err
	}

	return rebuildReadmeWith(files)
}

func rebuildReadmeWith(files []string) error {
	templateContents := readmeContents
	temp, err := os.ReadFile("./adr/templates/template_readme.md")
	if err == nil {
		templateContents = string(temp)
	}

	formattedFiles := make([]string, 0, len(files))
	for _, str := range files {
		newStr := fmt.Sprintf(" - [%v](./%v)", str, str)
		formattedFiles = append(formattedFiles, newStr)
	}

	output := strings.NewReplacer(
		"{{timestamp}}", time.Now().UTC().Format(time.RFC1123),
		"{{contents}}", strings.Join(formattedFiles, "\n"),
	).Replace(templateContents)

	return os.WriteFile("./adr/README.md", []byte(output), 0666)
}

func generateAdr(n int, name string) error {
	templateContents := adrContents
	temp, err := os.ReadFile("./adr/templates/template_adr.md")
	if err == nil {
		templateContents = string(temp)
	}

	heading := fmt.Sprintf("%05v - %v", n, name)
	contents := strings.ReplaceAll(templateContents, "{{name}}", heading)

	fileName := fmt.Sprintf("./adr/%05v-%v.md", n, createSafeName(name, ' '))
	return os.WriteFile(fileName, []byte(contents), 0666)
}

func createSafeName(name string, replacement rune) string {
	var sb strings.Builder
	sb.Grow(len(name))
	for _, r := range name {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			sb.WriteRune(replacement)
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func establishCoreFiles() error {
	if err := os.MkdirAll("./adr/assets", 0666); err != nil {
		return err
	}
	if err := os.MkdirAll("./adr/templates", 0666); err != nil {
		return err
	}

	return os.WriteFile("./adr/templates/README.md", []byte(templateReadmeContents), 0666)
}

func getAllFilesInAdrDir() ([]string, error) {
	dir, err := os.ReadDir("adr")
	if err != nil {
		return nil, err
	}

	var fileList []string
	for _, entry := range dir {
		switch entry.Name() {
		case "README.md", "assets", "templates":
			continue
		}
		fileList = append(fileList, entry.Name())
	}
	return fileList, nil
}

func assertNoError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println(helpContents)
		return
	}

	switch cmd := os.Args[1]; cmd {
	case "create":
		name := strings.Join(os.Args[2:], " ")
		if name == "" {
			err := fmt.Errorf("no name supplied for the ADR.\n " +
				"Command should be: `adl create <Name of ADR here>`")
			assertNoError(err)
		}

		assertNoError(establishCoreFiles())
		fileList, err := getAllFilesInAdrDir()
		assertNoError(err)
		assertNoError(generateAdr(len(fileList), name))
		assertNoError(rebuildReadmeWith(fileList))

	case "regen":
		assertNoError(establishCoreFiles())
		assertNoError(rebuildReadme())

	default:
		err := fmt.Errorf("unknown command '%v'\n\n%v", cmd, helpContents)
		assertNoError(err)
	}
}
