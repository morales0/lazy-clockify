/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func parseHourMinute(s string) (hour, min int, err error) {
	_, err = fmt.Sscanf(s, "%d:%d", &hour, &min)
	return
}

func getGitBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func extractTicketNumber(branchName, prefix string) string {
	// Create regex pattern: prefix followed by dash and digits
	// e.g. EL-\d+
	pattern := fmt.Sprintf(`%s-\d+`, regexp.QuoteMeta(prefix))
	re := regexp.MustCompile(pattern)
	match := re.FindString(branchName)
	return match
}

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

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new time entry",
	Long:  `Add a new time entry to Clockify using configuration for start, end, date, and message.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
			return fmt.Errorf("Failed to get ticket number: %w", err)
		}

		fmt.Printf("Creating a new time entry from %s to %s for ticket %s\n",
			startTime.Local().Format("03:04 PM"),
			endTime.Local().Format("03:04 PM"),
			ticketNumber)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().String("start_time", "9:00", "Start time")
	newCmd.Flags().String("end_time", "17:00", "End time")
	newCmd.Flags().String("ticket_prefix", "JIRA", "Ticket Prefix")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
