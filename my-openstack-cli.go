package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const configFileName = "config"

func setActiveProject(projectID, projectName string) error {
	data := []byte(fmt.Sprintf("%s\n%s", projectID, projectName))
	return ioutil.WriteFile(configFileName, data, 0644)
}

func getActiveProject() (string, string, error) {
	data, err := ioutil.ReadFile(configFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil
		}
		return "", "", err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) >= 2 {
		projectID := strings.TrimSpace(lines[0])
		projectName := strings.TrimSpace(lines[1])
		return projectID, projectName, nil
	}
	return "", "", fmt.Errorf("Invalid format in config file")
}

func getProjectName(projectID string, projects map[string]string) (string, error) {
	projectName, found := projects[projectID]
	if !found {
		return "", fmt.Errorf("Project ID not found: %s", projectID)
	}
	return projectName, nil
}

func getAvailableProjects() (map[string]string, error) {
	cmd := exec.Command("openstack", "project", "list", "-f", "value", "-c", "ID", "-c", "Name")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	projects := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			projectID := fields[0]
			projectName := fields[1]
			projects[projectID] = projectName
		}
	}
	return projects, nil
}

func runOpenStackCommand(args []string) error {
	projectID, _, err := getActiveProject()
	if err != nil {
		return err
	}
	if projectID == "" {
		fmt.Println("No active project set.")
		return nil
	}

	openstackArgs := append([]string{"openstack"}, args...)
	openstackArgs = append(openstackArgs, "--os-project-id="+projectID)

	cmd := exec.Command(openstackArgs[0], openstackArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func printHelp() {
	helpText := `
Usage: my-openstack-cli <command> [<args>]

Commands:
  set-project <project-id>   Set the active project
  show-current                Display the active project
  show-projects              Display all available projects
  --help                     Show this help message
  <openstack-command>        Run an OpenStack CLI command with the active project

Example:
  my-openstack-cli set-project my-project-id
  my-openstack-cli show-current
  my-openstack-cli show-projects
  my-openstack-cli server list
`
	fmt.Println(helpText)
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" {
		printHelp()
		return
	}

	command := os.Args[1]
	switch command {
	case "set-project":
		if len(os.Args) < 4 {
			fmt.Println("Usage: my-openstack-cli set-project <project-id> <project-name>")
			return
		}
		projectID := os.Args[2]
		projectName := os.Args[3]
		err := setActiveProject(projectID, projectName)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Project set to:", projectID)

	case "show-current":
		projects, err := getAvailableProjects()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		projectID, _, err := getActiveProject()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if projectID == "" {
			fmt.Println("No active project set.")
			return
		}
		projectName, err := getProjectName(projectID, projects)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("Active project ID: %s\n", projectID)
		fmt.Printf("Active project Name: %s\n", projectName)
	
	case "show-projects":
		projects, err := getAvailableProjects()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Available projects:")
		for projectID, projectName := range projects {
			fmt.Printf("%s - %s\n", projectID, projectName)
		}
	
	default:
		args := os.Args[1:]
		err := runOpenStackCommand(args)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}	