package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mjavint/ocli/pkg/config"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
func NewConfigAddonCmd() *cobra.Command {
	var (
		addons []string
	)
	cmd := &cobra.Command{
		Use:   "addon",
		Short: "Configuration Addon in projects",
		Long:  `Configuration Addon in projects`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(addons) == 0 {
				addons = config.AppConfig.Odoo.Addons
			}
			addonsPath := strings.Join(addons, ",")
			if err := updateOdooConf(config.AppConfig.Odoo.ConfigFile, addonsPath); err != nil {
				log.Fatal("failed to update odoo.conf: %w", err)
			}
			// Obtener directorio actual (funciona en todos los SO)
			dir, err := os.Getwd()
			if err != nil {
				panic(fmt.Sprintf("Error obteniendo directorio actual: %v", err))
			}
			pyrightConfigPath := filepath.Join(dir, "pyrightconfig.json")
			if err := updatePyrightConfig(pyrightConfigPath, addonsPath); err != nil {
				log.Fatal("failed to update pyrightconfig.json: %w", err)
			}
			fmt.Printf("📦 Addons paths detectados (%d total):\n", len(addons))
			fmt.Printf("Successfully created %s\n", addonsPath)
		},
	}
	return cmd
}

// updateOdooConf updates the addons_path in odoo.conf file
func updateOdooConf(odooConfPath, newAddonsPath string) error {
	content, err := os.ReadFile(odooConfPath)
	if err != nil {
		return fmt.Errorf("failed to read odoo.conf: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	updated := false

	// Update existing addons_path or mark insertion point
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "addons_path") {
			lines[i] = fmt.Sprintf("addons_path = %s", newAddonsPath)
			updated = true
			break
		}
	}

	// If not found, append to end
	if !updated {
		lines = append(lines, fmt.Sprintf("addons_path = %s", newAddonsPath))
	}

	newContent := strings.Join(lines, "\n")
	return os.WriteFile(odooConfPath, []byte(newContent), 0644)
}

// updatePyrightConfig updates the extraPaths array in pyrightconfig.json
func updatePyrightConfig(pyrightConfigPath, newAddonsPath string) error {
	content, err := os.ReadFile(pyrightConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read pyrightconfig.json: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(content, &config); err != nil {
		return fmt.Errorf("failed to parse pyrightconfig.json: %w", err)
	}

	// Convert comma-separated addons path to slice
	addonPaths := strings.Split(newAddonsPath, ",")
	cleanPaths := make([]string, 0, len(addonPaths))
	for _, p := range addonPaths {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			cleanPaths = append(cleanPaths, trimmed)
		}
	}

	// Update extraPaths
	config["extraPaths"] = cleanPaths

	// Write back with indentation
	newContent, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal pyrightconfig.json: %w", err)
	}

	return os.WriteFile(pyrightConfigPath, newContent, 0644)
}
