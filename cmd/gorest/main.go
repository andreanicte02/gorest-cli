// Command gorest is a CLI for managing users on the GoRest API.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/andreanicte02/gorest-cli/gorest"
)

var version = "dev" // set by goreleaser

const usage = `gorest - manage GoRest (https://gorest.co.in) users from the terminal

Usage:
  gorest <command> [flags]

Commands:
  list     [--page N] [--per-page N]
  get      <id>
  create   --name NAME --email EMAIL --gender male|female [--status active|inactive]
  update   <id> [--name NAME] [--email EMAIL] [--gender G] [--status S]
  delete   <id>
  version

Write commands need a token from https://gorest.co.in/my-account/access-tokens,
passed with --token or the GOREST_TOKEN environment variable.
`

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stdout, usage)
		return nil
	}
	cmd, args := args[0], args[1:]

	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	token := fs.String("token", os.Getenv("GOREST_TOKEN"), "API token (default $GOREST_TOKEN)")
	page := fs.Int("page", 1, "page number")
	perPage := fs.Int("per-page", 10, "results per page")
	var u gorest.User
	fs.StringVar(&u.Name, "name", "", "user name")
	fs.StringVar(&u.Email, "email", "", "user email")
	fs.StringVar(&u.Gender, "gender", "", "male | female")
	fs.StringVar(&u.Status, "status", "", "active | inactive")

	// Allow the ID to come before the flags: `gorest update 123 --status inactive`.
	var id int
	if len(args) > 0 && args[0] != "" && args[0][0] != '-' {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid id %q", args[0])
		}
		id, args = n, args[1:]
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	c := gorest.New(*token)

	needID := func() error {
		if id == 0 {
			return fmt.Errorf("%s: missing user id", cmd)
		}
		return nil
	}
	needToken := func() error {
		if c.Token == "" {
			return errors.New(cmd + ": a token is required (use --token or GOREST_TOKEN)")
		}
		return nil
	}

	switch cmd {
	case "list":
		users, err := c.ListUsers(ctx, *page, *perPage)
		if err != nil {
			return err
		}
		return printJSON(stdout, users)

	case "get":
		if err := needID(); err != nil {
			return err
		}
		user, err := c.GetUser(ctx, id)
		if err != nil {
			return err
		}
		return printJSON(stdout, user)

	case "create":
		if err := needToken(); err != nil {
			return err
		}
		if u.Status == "" {
			u.Status = "active"
		}
		user, err := c.CreateUser(ctx, u)
		if err != nil {
			return err
		}
		return printJSON(stdout, user)

	case "update":
		if err := errors.Join(needID(), needToken()); err != nil {
			return err
		}
		user, err := c.UpdateUser(ctx, id, u)
		if err != nil {
			return err
		}
		return printJSON(stdout, user)

	case "delete":
		if err := errors.Join(needID(), needToken()); err != nil {
			return err
		}
		if err := c.DeleteUser(ctx, id); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "user %d deleted\n", id)
		return nil

	case "version":
		fmt.Fprintln(stdout, "gorest", version)
		return nil
	}
	return fmt.Errorf("unknown command %q (run 'gorest help')", cmd)
}

func printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
