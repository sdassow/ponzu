package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [flags] <project name>",
	Short: "creates a project directory of the name supplied as a parameter",
	Long: `Creates a project directory of the name supplied as a parameter
immediately following the 'new' option in the $GOPATH/src directory. Note:
'new' depends on the program 'git' and possibly a network connection. If
there is no local repository to clone from at the local machine's $GOPATH,
'new' will attempt to clone the 'github.com/sdassow/ponzu' package from
over the network.`,
	Example: `$ ponzu new github.com/nilslice/proj
> New ponzu project created at $GOPATH/src/github.com/nilslice/proj`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := "ponzu"
		if len(args) > 0 {
			projectName = args[0]
		} else {
			msg := "Please provide a project name."
			msg += "\nThis will create a directory within your $GOPATH/src."
			return fmt.Errorf("%s", msg)
		}
		return newProjectInDir(projectName)
	},
}

// name2path transforns a project name to an absolute path
func name2path(projectName string) (string, error) {
	
	path := projectName
	
	path = filepath.Join(".", path)
	

	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err == nil {
		err = os.ErrExist
	} else if os.IsNotExist(err) {
		err = nil
	}

	return path, err
}

func newProjectInDir(path string) error {
	prjname:=path
	path, err := name2path(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// path exists, ask if it should be overwritten
	if os.IsNotExist(err) {
		fmt.Printf("Using '%s' as project directory\n", path)
		fmt.Println("Path exists, overwrite contents? (y/N):")

		answer, err := getAnswer()
		if err != nil {
			return err
		}

		switch answer {
		case "n", "no", "\r\n", "\n", "":
			fmt.Println("")

		case "y", "yes":
			err := os.RemoveAll(path)
			if err != nil {
				return fmt.Errorf("Failed to overwrite %s. \n%s", path, err)
			}

			return createProjectInDir(path,prjname)

		default:
			fmt.Println("Input not recognized. No files overwritten. Answer as 'y' or 'n' only.")
		}

		return nil
	}

	return createProjectInDir(path,prjname)
}

func createProjectInDir(path string,prjname string) error {
	
	repo := ponzuRepo
	network := "https://" + strings.Join(repo, "/") + ".git"
	

	// create the directory or overwrite it
	err := os.MkdirAll(path, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}
	
		networkClone := exec.Command("git", "clone", network, path)
		networkClone.Stdout = os.Stdout
		networkClone.Stderr = os.Stderr

		err = networkClone.Start()
		if err != nil {
			fmt.Println("Network clone failed to start. Try again and make sure you have a network connection.")
			return err
		}
		err = networkClone.Wait()
		if err != nil {
			fmt.Println("Network clone failure.")
			// failed
			return fmt.Errorf("Failed to clone files over the network [%s].\n%s", network, err)
		}


	// remove non-project files and directories
	rmPaths := []string{".git", ".circleci","system"}
	for _, rm := range rmPaths {
		dir := filepath.Join(path, rm)
		err = os.RemoveAll(dir)
		if err != nil {
			fmt.Println("Failed to remove directory from your project path. Consider removing it manually:", dir)
		}
	}

	content,err1:=os.ReadFile(filepath.Join(path,"go_dev.mod"))
	if err1!=nil {
		return fmt.Errorf("Failed to read go_dev.mod %s.\n", err1)
	}
	ncontent:=strings.ReplaceAll(string(content),"##MODNAME##",prjname)
	err=os.WriteFile(filepath.Join(path,"go.mod"),[]byte(ncontent),0666)
	if err!=nil {
		return fmt.Errorf("Failed to write go.mod %s.\n", err)
	}
	content,err1=os.ReadFile(filepath.Join(path,"cmd","ponzu","main_dev"))
	if err1!=nil {
		return fmt.Errorf("Failed to read main_dev.mod %s.\n", err1)
	}
	ncontent=strings.ReplaceAll(string(content),"##MODNAME##",prjname)
	err=os.WriteFile(filepath.Join(path,"cmd","ponzu","main.go"),[]byte(ncontent),0666)
	if err!=nil {
		return fmt.Errorf("Failed to write go.mod %s.\n", err)
	}
	fmt.Println("New ponzu project created at", path)
	return nil
}

func init() {
	RegisterCmdlineCommand(newCmd)
}
