/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/socialviolation/asciiban/ascii"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mcpv",
	Short: "MCP Server Version Manager",
	Long: `mcpv is a CLI tool for managing Model Context Protocol (MCP) servers.

It allows you to install, update, and delete MCP servers on your local machine,
with support for multiple versions of the same server. Each project can have
its own mcpv.json configuration file to specify required server versions.

Examples:
  mcpv install                    # Install servers from mcpv.json
  mcpv install server@1.0.0       # Install specific server version
  mcpv list                       # List installed servers
  mcpv update server              # Update server to latest version
  mcpv remove server@1.0.0        # Remove specific server version`,
	Run: func(cmd *cobra.Command, args []string) {
		ascii.Draw(
			ascii.WithMessage("mcpv."),
			ascii.WithFont(ascii.FontUnivers),
			ascii.WithPalette(ascii.PaletteLime),
		)
		fmt.Printf("Version: %s\nCommit: %s\n", Version, Commit)
		_ = cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
}
