package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/BioWare/lazyobsidian/internal/config"
	"github.com/BioWare/lazyobsidian/internal/logging"
	"github.com/BioWare/lazyobsidian/internal/ui"
)

var (
	version   = "0.1.0"
	cfgFile   string
	vaultPath string
	themeName string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "lazyobsidian",
		Short: "TUI productivity dashboard for Obsidian vault",
		Long: `LazyObsidian is a terminal-based productivity dashboard
that works with your Obsidian vault. It provides keyboard-first
navigation, pomodoro timer, goal tracking, and more.`,
		RunE: run,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/lazyobsidian/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&vaultPath, "vault", "", "path to Obsidian vault")
	rootCmd.PersistentFlags().StringVar(&themeName, "theme", "", "theme to use (corsair-light, corsair-dark)")

	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Initialize logging (always enabled for now)
	if err := logging.Init(true); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize logging: %v\n", err)
	}
	defer logging.Close()

	logging.Info("Starting LazyObsidian v%s", version)

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		logging.Error("Failed to load config: %v", err)
		return fmt.Errorf("failed to load config: %w", err)
	}
	logging.Info("Config loaded successfully")

	// Override vault path if provided via flag
	if vaultPath != "" {
		cfg.Vault.Path = vaultPath
		logging.Info("Vault path overridden via flag: %s", vaultPath)
	}

	// Override theme if provided via flag
	if themeName != "" {
		cfg.Theme.Current = themeName
		logging.Info("Theme overridden via flag: %s", themeName)
	}

	// Validate vault path
	if cfg.Vault.Path == "" {
		logging.Error("Vault path is empty")
		return fmt.Errorf("vault path is required. Use --vault flag or set it in config file")
	}

	// Expand home directory
	cfg.Vault.Path = expandPath(cfg.Vault.Path)
	logging.Info("Using vault path: %s", cfg.Vault.Path)

	// Check if vault exists
	if _, err := os.Stat(cfg.Vault.Path); os.IsNotExist(err) {
		logging.Error("Vault path does not exist: %s", cfg.Vault.Path)
		return fmt.Errorf("vault path does not exist: %s", cfg.Vault.Path)
	}

	// Start the TUI application
	logging.Info("Starting TUI...")
	return ui.Run(cfg)
}

func loadConfig() (*config.Config, error) {
	if cfgFile != "" {
		return config.LoadFromFile(cfgFile)
	}
	return config.Load()
}

func expandPath(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("lazyobsidian version %s\n", version)
		},
	}
}
