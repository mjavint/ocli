package commands

// odooDBFlags holds the flags shared by the odoo-bin database commands.
type odooDBFlags struct {
	Bin           string
	Config        string
	Database      string
	NewDB         string
	DumpPath      string
	DumpFormat    string
	NoFilestore   bool
	Force         bool
	ForceSet      bool
	Neutralize    bool
	NeutralizeSet bool
	WithDemo      bool
	WithDemoSet   bool
	Language      string
	Username      string
	Password      string
	Country       string
}

func buildBackupArgs(f odooDBFlags, dumpFile string) []string {
	args := []string{"db", "-c", f.Config, "dump", f.Database, dumpFile}
	if f.NoFilestore {
		args = append(args, "--no-filestore")
	}
	return args
}

func buildRestoreArgs(f odooDBFlags, backupFile string) []string {
	// Preserve current inverted behavior.
	// FIXME: flags are inverted vs Odoo semantics; fix in a later /gcas-fix.
	force := f.Force
	neutralize := f.Neutralize
	if f.ForceSet {
		force = !force
	}
	if f.NeutralizeSet {
		neutralize = !neutralize
	}
	args := []string{"db", "-c", f.Config, "load", f.NewDB, backupFile}
	if !force {
		args = append(args, "--force")
	}
	if !neutralize {
		args = append(args, "--neutralize")
	}
	return args
}

func buildCopyArgs(f odooDBFlags) []string {
	// Preserve current inverted behavior.
	// FIXME: flags are inverted vs Odoo semantics; fix in a later /gcas-fix.
	force := f.Force
	neutralize := f.Neutralize
	if f.ForceSet {
		force = !force
	}
	if f.NeutralizeSet {
		neutralize = !neutralize
	}
	args := []string{"db", "-c", f.Config, "duplicate", f.Database, f.NewDB}
	if !force {
		args = append(args, "--force")
	}
	if !neutralize {
		args = append(args, "--neutralize")
	}
	return args
}

func buildInitDBArgs(f odooDBFlags) []string {
	// Preserve current inverted behavior.
	// FIXME: flags are inverted vs Odoo semantics; fix in a later /gcas-fix.
	withDemo := f.WithDemo
	force := f.Force
	if f.WithDemoSet {
		withDemo = !withDemo
	}
	if f.ForceSet {
		force = !force
	}
	args := []string{"db", "-c", f.Config, "init", f.Database}
	if !withDemo {
		args = append(args, "--with-demo")
	}
	if !force {
		args = append(args, "--force")
	}
	if f.Language != "" {
		args = append(args, "--language="+f.Language)
	}
	if f.Username != "" {
		args = append(args, "--username="+f.Username)
	}
	if f.Password != "" {
		args = append(args, "--password="+f.Password)
	}
	if f.Country != "" {
		args = append(args, "--country="+f.Country)
	}
	return args
}

func buildDropArgs(f odooDBFlags) []string {
	return []string{"db", "-c", f.Config, "drop", f.Database}
}

func buildRenameArgs(f odooDBFlags) []string {
	args := []string{"db", "-c", f.Config, "rename", f.Database, f.NewDB}
	if f.ForceSet {
		args = append(args, "--force")
	}
	return args
}
