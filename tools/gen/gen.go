package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const BASE_DIR_NAME = "revit"
const MODULE_ROOT = "revit"
const MANIFEST_FILENAME = "tools/manifest/gen_manifest.txt"

func loadManifest() ([]string) {
	f, err := os.OpenFile(MANIFEST_FILENAME, os.O_RDONLY, 0744)
	if err != nil {
		ExitFail(4, "Could not read the manifest file. Make sure you have the proper permissions and the file exists.", func(){})
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		ExitFail(4, "Could not read the manifest file. Make sure you have the proper permissions and the file exists.", func(){})
	}
	content := string(b)

	entries := strings.Split(content, ";")
	output := make([]string, 0, len(entries))
	for _, e := range entries {
		if !strings.Contains(e, "/") {
			continue
		}
		tmpOut := strings.ReplaceAll(e, "\n", "")
		output = append(output, strings.ReplaceAll(tmpOut, " ", ""))
	}
	fmt.Println(output)
	return output
}

func ExitFail(code int, msg string, cleanup func()) {
	cleanup()
	fmt.Printf("[FATAL ERROR]: %s\n> exiting...\n", msg)
	os.Exit(code)
}

func main() {
	dirName, err := os.Getwd()
	if err != nil {
		ExitFail(2, "Could not read the working directory. Make sure you have the proper permissions.", func(){})
	}
	dirNameParts := strings.Split(dirName, "/")
	currentDir := dirNameParts[len(dirNameParts) - 1]
	if currentDir != BASE_DIR_NAME {
		ExitFail(3, fmt.Sprintf("Please run the script in %s, the project root. Directory %s detected instead.", BASE_DIR_NAME, currentDir), func(){})
	}
	targets := loadManifest()
	for _, target := range targets {
		module := fmt.Sprint(MODULE_ROOT, target)
		cmd := exec.Command("go", "generate", module)
		out, err := cmd.Output()
		if err != nil {
			ExitFail(5, fmt.Sprintf("`go generate %s` failed with output %s and error %s", module, string(out), cmd.Err), func(){})
		}
		fmt.Printf("Code generation for module %s was successful\n", module)
	}
}
