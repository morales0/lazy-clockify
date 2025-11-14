package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/morales0/lazy-clockify/clockify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure workspace and project for time entries",
	Long: `Interactively configure your default workspace and project.
This command will query available workspaces, let you choose one,
then query projects for that workspace and let you choose a default project.
The configuration will be saved to your config file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get API key from config
		apiKey := viper.GetString("api_key")
		if apiKey == "" {
			return fmt.Errorf("api_key not found in configuration. Please set it in your config file first")
		}

		// Create Clockify client
		client := clockify.NewClient(apiKey)

		// Get workspaces
		fmt.Println("Fetching available workspaces...")
		workspaces, err := client.GetWorkspaces()
		if err != nil {
			return fmt.Errorf("failed to get workspaces: %w", err)
		}

		if len(workspaces) == 0 {
			return fmt.Errorf("no workspaces found for this user")
		}

		// Display workspaces
		fmt.Println("\nAvailable Workspaces:")
		fmt.Println(strings.Repeat("=", 60))
		for i, ws := range workspaces {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, ws.Name, ws.ID)
		}
		fmt.Println(strings.Repeat("=", 60))

		// Prompt for workspace selection
		reader := bufio.NewReader(os.Stdin)
		var selectedWorkspace *clockify.Workspace
		for {
			fmt.Printf("\nSelect a workspace (1-%d): ", len(workspaces))
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			input = strings.TrimSpace(input)
			choice, err := strconv.Atoi(input)
			if err != nil || choice < 1 || choice > len(workspaces) {
				fmt.Printf("Invalid choice. Please enter a number between 1 and %d.\n", len(workspaces))
				continue
			}

			selectedWorkspace = &workspaces[choice-1]
			break
		}

		fmt.Printf("\n✓ Selected workspace: %s\n", selectedWorkspace.Name)

		// Get projects for the selected workspace
		fmt.Println("\nFetching available projects...")
		projects, err := client.GetProjects(selectedWorkspace.ID)
		if err != nil {
			return fmt.Errorf("failed to get projects: %w", err)
		}

		if len(projects) == 0 {
			return fmt.Errorf("no projects found in workspace '%s'", selectedWorkspace.Name)
		}

		// Display projects
		fmt.Println("\nAvailable Projects:")
		fmt.Println(strings.Repeat("=", 60))
		for i, proj := range projects {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, proj.Name, proj.ID)
		}
		fmt.Println(strings.Repeat("=", 60))

		// Prompt for project selection
		var selectedProject *clockify.Project
		for {
			fmt.Printf("\nSelect a default project (1-%d): ", len(projects))
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			input = strings.TrimSpace(input)
			choice, err := strconv.Atoi(input)
			if err != nil || choice < 1 || choice > len(projects) {
				fmt.Printf("Invalid choice. Please enter a number between 1 and %d.\n", len(projects))
				continue
			}

			selectedProject = &projects[choice-1]
			break
		}

		fmt.Printf("\n✓ Selected project: %s\n", selectedProject.Name)

		// Save to config
		viper.Set("workspace_id", selectedWorkspace.ID)
		viper.Set("project_id", selectedProject.ID)

		// Write config to file
		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			// If no config file exists, create one in the current directory
			configFile = "./config.yaml"
		}

		if err := viper.WriteConfigAs(configFile); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("Configuration saved successfully!")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Printf("Workspace:     %s (%s)\n", selectedWorkspace.Name, selectedWorkspace.ID)
		fmt.Printf("Project:       %s (%s)\n", selectedProject.Name, selectedProject.ID)
		fmt.Printf("Config file:   %s\n", configFile)
		fmt.Println(strings.Repeat("=", 60))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
