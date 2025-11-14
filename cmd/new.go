/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/morales0/lazy-clockify/clockify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func parseHourMinute(s string) (hour, min int, err error) {
	_, err = fmt.Sscanf(s, "%d:%d", &hour, &min)
	return
}

// getGitBranch returns the current git branch name, or an error if not in a git repo
func getGitBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// extractTicketNumber extracts a ticket number from a branch name based on the prefix
// e.g. "feature/EL-1234-some-description" with prefix "EL" returns "EL-1234"
func extractTicketNumber(branchName, prefix string) string {
	// Create regex pattern: prefix followed by dash and digits
	// e.g. EL-\d+
	pattern := fmt.Sprintf(`%s-\d+`, regexp.QuoteMeta(prefix))
	re := regexp.MustCompile(pattern)
	match := re.FindString(branchName)
	return match
}

// getTicketNumber tries to get ticket number from git branch, or prompts user
func getTicketNumber() (string, error) {
	prefix := viper.GetString("ticket_prefix")

	// Try to get ticket from git branch
	branchName, err := getGitBranch()
	if err == nil && branchName != "" {
		ticket := extractTicketNumber(branchName, prefix)
		if ticket != "" {
			fmt.Printf("Found ticket number from branch: %s\n", ticket)
			return ticket, nil
		}
		fmt.Printf("No ticket number found in branch: %s\n", branchName)
	} else {
		fmt.Println("Not in a git repository or no branch detected")
	}

	// Prompt user for ticket number
	fmt.Printf("Please enter ticket number (e.g. %s-1234): ", prefix)
	reader := bufio.NewReader(os.Stdin)
	ticket, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read ticket number: %w", err)
	}

	ticket = strings.TrimSpace(ticket)
	if ticket == "" {
		return "", fmt.Errorf("ticket number cannot be empty")
	}

	return ticket, nil
}

// formatDuration formats a duration in a human-readable format
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new time entry",
	Long:  `Add a new time entry to Clockify using configuration for start, end, date, and message.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get API key from config
		apiKey := viper.GetString("api_key")
		if apiKey == "" {
			return fmt.Errorf("api_key not found in configuration. Please set it in your config file or use --api_key flag")
		}

		// Get start and end time
		startTimeStr := viper.GetString("start_time")
		endTimeStr := viper.GetString("end_time")

		startHour, startMin, err := parseHourMinute(startTimeStr)
		if err != nil {
			return fmt.Errorf("invalid start_time format: %w", err)
		}
		endHour, endMin, err := parseHourMinute(endTimeStr)
		if err != nil {
			return fmt.Errorf("invalid end_time format: %w", err)
		}

		now := time.Now()
		startTime := time.Date(now.Year(), now.Month(), now.Day(), startHour, startMin, 0, 0, now.Location())
		endTime := time.Date(now.Year(), now.Month(), now.Day(), endHour, endMin, 0, 0, now.Location())

		// Get ticket number
		ticketNumber, err := getTicketNumber()
		if err != nil {
			return fmt.Errorf("failed to get ticket number: %w", err)
		}

		// Create Clockify client
		client := clockify.NewClient(apiKey)

		// Get user info to retrieve default workspace
		fmt.Println("Fetching user information...")
		user, err := client.GetUser()
		if err != nil {
			return fmt.Errorf("failed to get user info: %w", err)
		}

		workspaceID := user.DefaultWorkspace
		if workspaceID == "" {
			// Fallback: get the first workspace
			fmt.Println("No default workspace found, fetching workspaces...")
			workspaces, err := client.GetWorkspaces()
			if err != nil {
				return fmt.Errorf("failed to get workspaces: %w", err)
			}
			if len(workspaces) == 0 {
				return fmt.Errorf("no workspaces found for this user")
			}
			workspaceID = workspaces[0].ID
			fmt.Printf("Using workspace: %s\n", workspaces[0].Name)
		}

		// Get project id from workspace
		projects, err := client.GetProjects(workspaceID)
		if err != nil {
			return fmt.Errorf("failed to get projects: %w", err)
		}
		if len(projects) == 0 {
			return fmt.Errorf("no projects found for this user")
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
		reader := bufio.NewReader(os.Stdin)
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

		// Create time entry
		description := ticketNumber
		entryRequest := clockify.TimeEntryRequest{
			Start:       startTime.UTC(),
			End:         &endTime,
			Description: description,
			Billable:    true,
			ProjectID:   &selectedProject.ID,
		}

		// Display the request details
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("Time Entry Details")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Printf("Workspace ID:  %s\n", workspaceID)
		fmt.Printf("Project Name:  %s\n", selectedProject.Name)
		fmt.Printf("User:          %s (%s)\n", user.Name, user.Email)
		fmt.Printf("Ticket:        %s\n", ticketNumber)
		fmt.Printf("Description:   %s\n", description)
		fmt.Printf("Start Time:    %s (Local) / %s (UTC)\n",
			startTime.Local().Format("2006-01-02 03:04:05 PM MST"),
			startTime.UTC().Format("2006-01-02 15:04:05 UTC"))
		fmt.Printf("End Time:      %s (Local) / %s (UTC)\n",
			endTime.Local().Format("2006-01-02 03:04:05 PM MST"),
			endTime.UTC().Format("2006-01-02 15:04:05 UTC"))
		duration := endTime.Sub(startTime)
		fmt.Printf("Duration:      %s (%.2f hours)\n",
			formatDuration(duration),
			duration.Hours())
		fmt.Printf("Billable:      %t\n", entryRequest.Billable)
		fmt.Println(strings.Repeat("=", 60))

		// Show JSON payload
		fmt.Println("\nAPI Request Payload:")
		jsonPayload, err := json.MarshalIndent(entryRequest, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		fmt.Println(string(jsonPayload))
		fmt.Printf("\nAPI Endpoint: POST https://api.clockify.me/api/v1/workspaces/%s/time-entries\n", workspaceID)
		fmt.Println(strings.Repeat("=", 60))

		// Prompt for confirmation
		fmt.Print("\nDo you want to submit this time entry? (yes/no): ")
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "yes" && response != "y" {
			fmt.Println("Time entry cancelled.")
			return nil
		}

		// Submit the time entry
		fmt.Println("\nSubmitting time entry...")
		timeEntry, err := client.CreateTimeEntry(workspaceID, entryRequest)
		if err != nil {
			return fmt.Errorf("failed to create time entry: %w", err)
		}

		fmt.Printf("\n✓ Time entry created successfully!\n")
		fmt.Printf("  Entry ID: %s\n", timeEntry.ID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().String("start_time", "9:00", "Start time")
	newCmd.Flags().String("end_time", "17:00", "End time")
	newCmd.Flags().String("ticket_prefix", "JIRA", "Prefix for git branch ticket numbers (e.g. EL for EL-1234)")
	// newCmd.Flags().String("message", "", "Optional message to append to the time entry description")
	// newCmd.Flags().String("api_key", "", "Clockify API key")

	// viper.BindPFlag("start_time", newCmd.Flags().Lookup("start_time"))
	// viper.BindPFlag("end_time", newCmd.Flags().Lookup("end_time"))
	// viper.BindPFlag("git_ticket_prefix", newCmd.Flags().Lookup("git_ticket_prefix"))
	// viper.BindPFlag("message", newCmd.Flags().Lookup("message"))
	// viper.BindPFlag("api_key", newCmd.Flags().Lookup("api_key"))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
