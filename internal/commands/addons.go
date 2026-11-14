package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// NewConfigAddonCmd creates the addon command.
func NewConfigAddonCmd(deps Deps) *cobra.Command {
	var addons []string
	cmd := &cobra.Command{
		Use:   "addon",
		Short: "Configuration Addon in projects",
		Long:  `Configuration Addon in projects`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := deps.cfg()
			if len(addons) == 0 {
				addons = cfg.Odoo.Addons
			}
			addonsPath := strings.Join(addons, ",")
			if err := updateOdooConf(cfg.Odoo.ConfigFile, addonsPath); err != nil {
				return fmt.Errorf("failed to update odoo.conf: %w", err)
			}
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("Error obteniendo directorio actual: %v", err)
			}
			pyrightConfigPath := filepath.Join(dir, "pyrightconfig.json")
			if err := updatePyrightConfig(pyrightConfigPath, addonsPath); err != nil {
				return fmt.Errorf("failed to update pyrightconfig.json: %w", err)
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "📦 Addons paths detectados (%d total):\n", len(addons))
			fmt.Fprintf(out, "Successfully created %s\n", addonsPath)
			return nil
		},
	}
	setCmdIO(cmd, deps)
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

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "addons_path") {
			lines[i] = fmt.Sprintf("addons_path = %s", newAddonsPath)
			updated = true
			break
		}
	}

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

	var cfg map[string]interface{}
	if err := json.Unmarshal(content, &cfg); err != nil {
		return fmt.Errorf("failed to parse pyrightconfig.json: %w", err)
	}

	addonPaths := strings.Split(newAddonsPath, ",")
	cleanPaths := make([]string, 0, len(addonPaths))
	for _, p := range addonPaths {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			cleanPaths = append(cleanPaths, trimmed)
		}
	}

	cfg["extraPaths"] = cleanPaths

	newContent, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal pyrightconfig.json: %w", err)
	}

	return os.WriteFile(pyrightConfigPath, newContent, 0644)
}
