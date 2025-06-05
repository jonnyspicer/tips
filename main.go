// Package main provides a CLI tool for displaying and managing helpful tips.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Version information
const Version = "1.0.0"

// Command-line flags
var (
	topicFlag   []string // Topics to filter tips by
	refreshFlag int      // Refresh interval in minutes
	countFlag   int      // Number of tips to generate per API call
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "tips",
	Short:   "A CLI tool for displaying helpful tips",
	Version: Version,
	Long: `Tips is a command-line tool that displays helpful tips on various topics.

It allows you to generate tips using LLM providers (OpenAI, Anthropic, Google),
store them locally, and display them in an interactive terminal interface.`,
	Run: func(cmd *cobra.Command, args []string) {
		showCmd.Run(cmd, args)
	},
}

// showCmd displays tips in an interactive terminal interface
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Display tips in interactive mode",
	Long: `Display tips in an interactive terminal interface.

Use 'n' to get next tip, 'k' to mark current tip as known, 'q' to quit.
Tips can be filtered by topic using the --topic flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		if refreshFlag <= 0 {
			fmt.Fprintf(os.Stderr, "Error: Refresh interval must be greater than 0 minutes\n")
			os.Exit(1)
		}
		if err := runBubbleTeaShow(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// generateCmd generates new tips using LLM providers
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate new tips for specified topics",
	Long: `Generate new tips using AI language models.

Requires API key for your chosen provider:
- OpenAI: Set OPENAI_API_KEY environment variable
- Anthropic: Set ANTHROPIC_API_KEY environment variable  
- Google: Set GOOGLE_API_KEY environment variable

Set model with TIPS_MODEL (default: openai/gpt-4o)
Format: provider/model (e.g., anthropic/claude-3-sonnet-20240229)`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(topicFlag) == 0 {
			fmt.Fprintf(os.Stderr, "Error: Please specify at least one topic using -t or --topic\n")
			os.Exit(1)
		}

		for _, topic := range topicFlag {
			topic = strings.TrimSpace(topic)
			if topic == "" {
				fmt.Fprintf(os.Stderr, "Warning: Empty topic provided, skipping\n")
				continue
			}

			fmt.Printf("Generating %d tips for topic: %s...\n", countFlag, topic)
			
			if countFlag <= 0 {
				fmt.Fprintf(os.Stderr, "Error: Count must be greater than 0, got %d\n", countFlag)
				continue
			}

			tips, err := generateTips(topic, countFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating tips for %s: %v\n", topic, err)
				continue
			}

			tipsData, err := loadTips()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading existing tips: %v\n", err)
				continue
			}

			for _, tip := range tips {
				tipsData.addTip(topic, tip.Content)
			}

			if err := saveTips(tipsData); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving tips: %v\n", err)
				continue
			}

			fmt.Printf("Successfully generated and saved %d tips for %s\n", len(tips), topic)
		}
	},
}

// clearCmd removes all stored tips
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Delete all stored tips",
	Long:  `Delete all tips from local storage (~/.tips.json).`,
	Run: func(cmd *cobra.Command, args []string) {
		filePath, err := getTipsFilePath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting tips file path: %v\n", err)
			os.Exit(1)
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Println("No tips file found - nothing to clear")
			return
		}

		if err := os.Remove(filePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting tips file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully deleted tips file: %s\n", filePath)
	},
}

func init() {
	rootCmd.PersistentFlags().StringSliceVarP(&topicFlag, "topic", "t", []string{}, "Filter by topic (can specify multiple)")
	rootCmd.PersistentFlags().IntVarP(&refreshFlag, "refresh", "r", 60, "Refresh interval in minutes")
	rootCmd.PersistentFlags().IntVarP(&countFlag, "count", "c", 20, "Number of tips to generate per API call")

	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(clearCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}