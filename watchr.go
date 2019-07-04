package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/urfave/cli"
)

func watchFile(file string, cmd string, quiet bool, verbose bool) error {
	if !quiet {
		fmt.Printf("*** Starting watchr for the file: %s\n", file)
	}

	// Get the watched file stats to store first modification date for comparison later
	inf, err := os.Stat(file)
	if err != nil {
		log.Fatal(err)
	}

	modTime := inf.ModTime()
	var modCount int            // Modifications counter for stats
	var diff time.Duration      // Difference between last known modification and current modification times
	var totalDiff time.Duration // Total time between all modifications for stats

	// Main loop, run the modification time comparison and command execution infinitely
	for {
		// Check the watched file stats again
		inf, err := os.Stat(file)
		if err != nil {
			log.Fatal(err)
		}

		diff = inf.ModTime().Sub(modTime) // Compare last known modTime to the current one
		if diff != 0 {                    // Time difference detected, file was modified
			modTime = inf.ModTime() // Save new modTime
			modCount++              // increase modification counter
			totalDiff += diff       // add new modification duration to total durations

			if !quiet {
				fmt.Printf("** The file %s was modified at: %s\n", file, inf.ModTime())
			}

			if cmd == "" { // If the --cmd flag was not set and --verbose was, print info
				if !quiet && verbose {
					fmt.Println("* Not executing any command")
					fmt.Printf("** Stats: %d modifications, last modified %s ago, average modification time %s\n",
						modCount, diff, totalDiff/time.Duration(modCount))
				}
			} else { // If the --cmd flag has been set, execute the command provided
				if !quiet {
					fmt.Printf("* Executing: %s\n", cmd)
				}

				s := strings.Fields(cmd)         // Split the cmd string into binary and arguments strings
				bin := s[0]                      // First part of the cmd string is the binary
				args := strings.Join(s[1:], " ") // Rest of the cmd string are the binary arguments, if any

				exe := exec.Command(bin, args) // Execute the command and store its output
				exeStart := time.Now()         // Store the time before command execution for measuring its execution time
				out, err := exe.Output()
				exeDur := time.Since(exeStart) // Store the execution time of the command for stats
				if err != nil {
					log.Fatal(err)
				}
				if !quiet && verbose {
					fmt.Printf("* Command output:\n%s", out)
					fmt.Printf("** Stats: %d modifications, last modified %s ago, average modification time %s, command execution %s\n",
						modCount, diff, totalDiff/time.Duration(modCount), exeDur)
				}
			}
		}
	}
}

func main() {
	// Watch for file modifications and execute given command when such modification
	// is detected.

	app := cli.NewApp()
	app.Name = "watchr"
	app.Version = "1.0.0"
	app.Usage = "Watch given file for modifications and execute commands when they are detected"

	var cmd, file string
	var quiet, verbose bool

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "cmd",
			Value:       "",
			Usage:       "Command to execute, when file modification is detected, eg. curl (optional)",
			Destination: &cmd,
		},
		cli.StringFlag{
			Name:        "file",
			Value:       "",
			Usage:       "Path to the file to watch for modifications, eg. foobar.go (required)",
			Destination: &file,
		},
		cli.BoolFlag{
			Name:        "quiet",
			Usage:       "Enable quiet operation and supress any and all output (optional, not usable with --verbose)",
			Destination: &quiet,
		},
		cli.BoolFlag{
			Name:        "verbose",
			Usage:       "Enable verbose output, including command execution output (optional, not usable with --quiet)",
			Destination: &verbose,
		},
	}

	app.Action = func(c *cli.Context) error {
		// Check if we have at least --file flag passed
		if c.NumFlags() < 1 {
			fmt.Println("Missing --file flag with file path")
			cli.ShowAppHelp(c)
			os.Exit(1)
		}
		if quiet && verbose {
			fmt.Println("The --quiet and --verbose flags are mutually exclusive")
			cli.ShowAppHelp(c)
			os.Exit(1)
		}
		// main application code
		err := watchFile(file, cmd, quiet, verbose)
		return err
	}
	// Run the application
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
