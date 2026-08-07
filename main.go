package main

import "os"
import "os/exec"
import "fmt"
import "path/filepath"
import "strings"
import "bufio"

const EDITOR string = "vi"

type FileDirectory struct {
	indexedAt       int
	directoryPath 	string
	orignalFileName string
	newFileName 		string
}

func main() {
	if len(os.Args) > 1 {

		directoryPath := os.Args[1] // Intentional: will only take the first arguement.

		dirContents, err := os.ReadDir(directoryPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fds := make([]FileDirectory, 0, len(dirContents))

		for i := 0; i < len(dirContents); i++ {
			// Ignore hidden files:
			if strings.HasPrefix(dirContents[i].Name(), ".") {
				continue
			}
			fileInfo, _ := os.Stat(filepath.Join(directoryPath, dirContents[i].Name()))
			if fileInfo.IsDir() {
				continue
			}

			fd := FileDirectory{
				indexedAt:       i,
				directoryPath: 	 directoryPath,
				orignalFileName: dirContents[i].Name(),
			}
			fds = append(fds, fd)
		}

		text := ""
		fmSaveFileLocation := "/tmp/fm_files_list"
		for i := 0; i < len(fds); i++ {
			text += fds[i].orignalFileName + "\n"
		}
		err = os.WriteFile(fmSaveFileLocation, []byte(text), 0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		cmd := exec.Command(EDITOR, fmSaveFileLocation)
		cmd.Stdin  = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Read back the changes:
		data, err := os.ReadFile(fmSaveFileLocation)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		rawLines := strings.Split(string(data), "\n")
		lines := make([]string, 0, len(rawLines))
		for _, line := range rawLines {
			if strings.TrimSpace(line) != "" {
				lines = append(lines, line)
			}
		}

		fmt.Println("Dry Run: these will be renamed:")
		if len(lines) == len(fds) {
			for i := 0; i < len(lines); i++ {
				fds[i].newFileName = lines[i]
				if fds[i].orignalFileName != fds[i].newFileName {
					fmt.Printf("%-40s -> %-40s\n",fds[i].orignalFileName, fds[i].newFileName)
				}
			}
		}

		fmt.Println("Confirm to apply changes N/y")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		input = strings.TrimSpace(input)
		if len(input) == 0 {
			input = "N" // Default response is "No" if the user doesn't pass any.
		}

		switch input {
		case "Y" , "y" , "Yes" , "yes":
			for i := 0; i < len(fds); i++ {
				if fds[i].orignalFileName != fds[i].newFileName {
					originalPath := filepath.Join(fds[i].directoryPath, fds[i].orignalFileName)
					newPath      := filepath.Join(fds[i].directoryPath, fds[i].newFileName)
					if _, err := os.Stat(newPath); err == nil {
						fmt.Printf("File '%s' cannot be renamed to '%s' because it already exists, skipping.\n", originalPath, newPath)
						continue
					}
					err := os.Rename(originalPath, newPath)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
				}
			}
			fmt.Println("Files renanmed.")
		default:
			fmt.Println("Operation cancelled")
		}
	}
}
