/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func parseHourMinute(s string) (hour, min int, err error) {
	_, err = fmt.Sscanf(s, "%d:%d", &hour, &min)
	return
}

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new time entry",
	Long:  `Add a new time entry to Clockify using configuration for start, end, date, and message.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		fmt.Printf("Creating a new time entry from %s to %s\n",
			startTime.Local().Format("03:04 PM"),
			endTime.Local().Format("03:04 PM"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().String("start_time", "9:00", "Start time")
	newCmd.Flags().String("end_time", "17:00", "End time")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
