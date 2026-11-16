package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"text/tabwriter"
	"time"

	"github.com/mjengofiti/api/internal/config"
	"github.com/mjengofiti/api/internal/database"
	"github.com/mjengofiti/api/migrations"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	var steps int = 1
	var dirPath string

	switch command {
	case "up":
		upCmd := flag.NewFlagSet("up", flag.ExitOnError)
		upCmd.StringVar(&dirPath, "dir", "", "Optional filesystem path to migrations directory (default: embedded)")
		_ = upCmd.Parse(os.Args[2:])
	case "down":
		downCmd := flag.NewFlagSet("down", flag.ExitOnError)
		downCmd.IntVar(&steps, "steps", 1, "Number of migrations to roll back")
		downCmd.StringVar(&dirPath, "dir", "", "Optional filesystem path to migrations directory (default: embedded)")
		_ = downCmd.Parse(os.Args[2:])
	case "status":
		statusCmd := flag.NewFlagSet("status", flag.ExitOnError)
		statusCmd.StringVar(&dirPath, "dir", "", "Optional filesystem path to migrations directory (default: embedded)")
		_ = statusCmd.Parse(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n\n", command)
		printUsage()
		os.Exit(1)
	}

	if err := run(command, steps, dirPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Mjengofiti Database Migration Runner")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  migrate <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up            Apply all pending migrations")
	fmt.Println("  down          Roll back the last applied migration")
	fmt.Println("  status        Display the status of all migrations")
	fmt.Println("  help          Display this help message")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -steps int    Number of migrations to roll back (for 'down', default: 1)")
	fmt.Println("  -dir string   Override embedded migrations with external directory")
}

func run(command string, steps int, dirPath string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	var fsys fs.FS
	if dirPath != "" {
		fsys = os.DirFS(dirPath)
		fmt.Printf("Using external migrations from: %s\n", dirPath)
	} else {
		fsys = migrations.FS
		fmt.Println("Using embedded canonical migrations (apps/api/migrations)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("Connecting to database at %s:%d/%s...\n", cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	pool, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("database connection failed (%s:%d/%s): %w",
			cfg.Database.Host, cfg.Database.Port, cfg.Database.Name, err)
	}
	defer pool.Close()

	migrator := database.NewMigrator(pool.Pool, fsys)

	switch command {
	case "up":
		applied, err := migrator.Up(ctx)
		if err != nil {
			return fmt.Errorf("migration up failed: %w", err)
		}
		if len(applied) == 0 {
			fmt.Println("Database is up to date. No pending migrations.")
		} else {
			fmt.Printf("Successfully applied %d migration(s):\n", len(applied))
			for _, v := range applied {
				fmt.Printf("  ✓ Version %03d\n", v)
			}
		}

	case "down":
		rolledBack, err := migrator.Down(ctx, steps)
		if err != nil {
			return fmt.Errorf("migration down failed: %w", err)
		}
		if len(rolledBack) == 0 {
			fmt.Println("No migrations to roll back.")
		} else {
			fmt.Printf("Successfully rolled back %d migration(s):\n", len(rolledBack))
			for _, v := range rolledBack {
				fmt.Printf("  ↺ Version %03d\n", v)
			}
		}

	case "status":
		statuses, err := migrator.Status(ctx)
		if err != nil {
			return fmt.Errorf("fetching migration status: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "VERSION\tNAME\tSTATUS\tAPPLIED AT\tCHECKSUM")
		for _, s := range statuses {
			state := "Pending"
			appliedAt := "-"
			chk := "-"
			if s.Applied {
				state = "Applied"
				if s.AppliedAt != nil {
					appliedAt = s.AppliedAt.Format("2006-01-02 15:04:05 UTC")
				}
				if len(s.Checksum) >= 8 {
					chk = s.Checksum[:8]
				}
			}
			fmt.Fprintf(w, "%03d\t%s\t%s\t%s\t%s\n",
				s.Version, s.Name, state, appliedAt, chk)
		}
		_ = w.Flush()
	}

	return nil
}
