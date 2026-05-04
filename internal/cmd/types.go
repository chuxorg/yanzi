package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func RunTypes(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: yanzi types list --json")
	}

	switch args[0] {
	case "list":
		return runTypesList(args[1:])
	default:
		return errors.New("usage: yanzi types list --json")
	}
}

func runTypesList(args []string) error {
	fs := flag.NewFlagSet("types list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	jsonFlag := fs.Bool("json", false, "render types as JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 || !*jsonFlag {
		return errors.New("usage: yanzi types list --json")
	}

	data, err := artifactTypeCatalogJSON()
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}
