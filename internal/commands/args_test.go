package commands

import (
	"reflect"
	"testing"
)

func TestBuildBackupArgs(t *testing.T) {
	f := odooDBFlags{
		Config:      "/etc/odoo.conf",
		Database:    "mydb",
		NoFilestore: false,
	}
	got := buildBackupArgs(f, "/dumps/mydb.zip")
	want := []string{"db", "-c", "/etc/odoo.conf", "dump", "mydb", "/dumps/mydb.zip"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildBackupArgs() = %#v, want %#v", got, want)
	}

	f.NoFilestore = true
	got = buildBackupArgs(f, "/dumps/mydb_no_fs.zip")
	want = []string{"db", "-c", "/etc/odoo.conf", "dump", "mydb", "/dumps/mydb_no_fs.zip", "--no-filestore"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildBackupArgs(no filestore) = %#v, want %#v", got, want)
	}
}

func TestBuildRestoreArgs(t *testing.T) {
	// Preserve current inverted behavior.
	// FIXME: flags are inverted vs Odoo semantics; fix in a later /gcas-fix.
	f := odooDBFlags{
		Config:        "/etc/odoo.conf",
		NewDB:         "newdb",
		Force:         true,
		Neutralize:    true,
		ForceSet:      false,
		NeutralizeSet: false,
	}
	got := buildRestoreArgs(f, "/dumps/old.zip")
	want := []string{"db", "-c", "/etc/odoo.conf", "load", "newdb", "/dumps/old.zip"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildRestoreArgs(defaults) = %#v, want %#v", got, want)
	}

	f.ForceSet = true
	f.NeutralizeSet = true
	got = buildRestoreArgs(f, "/dumps/old.zip")
	want = []string{"db", "-c", "/etc/odoo.conf", "load", "newdb", "/dumps/old.zip", "--force", "--neutralize"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildRestoreArgs(flags changed, inverted) = %#v, want %#v", got, want)
	}
}

func TestBuildCopyArgs(t *testing.T) {
	// Preserve current inverted behavior.
	// FIXME: flags are inverted vs Odoo semantics; fix in a later /gcas-fix.
	f := odooDBFlags{
		Config:        "/etc/odoo.conf",
		Database:      "src",
		NewDB:         "dst",
		Force:         true,
		Neutralize:    true,
		ForceSet:      false,
		NeutralizeSet: false,
	}
	got := buildCopyArgs(f)
	want := []string{"db", "-c", "/etc/odoo.conf", "duplicate", "src", "dst"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildCopyArgs(defaults) = %#v, want %#v", got, want)
	}

	f.ForceSet = true
	f.NeutralizeSet = true
	got = buildCopyArgs(f)
	want = []string{"db", "-c", "/etc/odoo.conf", "duplicate", "src", "dst", "--force", "--neutralize"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildCopyArgs(flags changed, inverted) = %#v, want %#v", got, want)
	}
}

func TestBuildInitDBArgs(t *testing.T) {
	// Preserve current inverted behavior.
	// FIXME: flags are inverted vs Odoo semantics; fix in a later /gcas-fix.
	f := odooDBFlags{
		Config:      "/etc/odoo.conf",
		Database:    "newdb",
		WithDemo:    true,
		Force:       true,
		WithDemoSet: false,
		ForceSet:    false,
		Language:    "es_ES",
		Username:    "admin",
		Password:    "secret",
		Country:     "CU",
	}
	got := buildInitDBArgs(f)
	want := []string{"db", "-c", "/etc/odoo.conf", "init", "newdb", "--language=es_ES", "--username=admin", "--password=secret", "--country=CU"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildInitDBArgs(defaults) = %#v, want %#v", got, want)
	}

	f.WithDemoSet = true
	f.ForceSet = true
	got = buildInitDBArgs(f)
	want = []string{"db", "-c", "/etc/odoo.conf", "init", "newdb", "--with-demo", "--force", "--language=es_ES", "--username=admin", "--password=secret", "--country=CU"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildInitDBArgs(flags changed, inverted) = %#v, want %#v", got, want)
	}
}

func TestBuildDropArgs(t *testing.T) {
	f := odooDBFlags{Config: "/etc/odoo.conf", Database: "olddb"}
	got := buildDropArgs(f)
	want := []string{"db", "-c", "/etc/odoo.conf", "drop", "olddb"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildDropArgs() = %#v, want %#v", got, want)
	}
}

func TestBuildRenameArgs(t *testing.T) {
	f := odooDBFlags{Config: "/etc/odoo.conf", Database: "old", NewDB: "new"}
	got := buildRenameArgs(f)
	want := []string{"db", "-c", "/etc/odoo.conf", "rename", "old", "new"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildRenameArgs(defaults) = %#v, want %#v", got, want)
	}

	f.ForceSet = true
	got = buildRenameArgs(f)
	want = []string{"db", "-c", "/etc/odoo.conf", "rename", "old", "new", "--force"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildRenameArgs(force set) = %#v, want %#v", got, want)
	}
}
