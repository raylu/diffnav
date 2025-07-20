package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"github.com/urfave/cli/v3"

	"github.com/dlvhdr/diffnav/pkg/ui"
)

func main() {
	cmd := &cli.Command{
		Usage: "A git diff pager with a file tree",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "unified",
				Aliases:     []string{"u"},
				Usage:       "side-by-side (default) or unified diff",
				HideDefault: true,
			},
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "log to debug.log",
				Sources: cli.EnvVars("DEBUG"),
			},
		},
		Action: runDiffNav,
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func runDiffNav(ctx context.Context, cmd *cli.Command) error {
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	if stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0 {
		fmt.Println("No diff, exiting")
		os.Exit(0)
	}

	if cmd.Bool("debug") {
		var fileErr error
		logFile, fileErr := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if fileErr != nil {
			fmt.Println("Error opening debug.log:", fileErr)
			os.Exit(1)
		}
		defer logFile.Close()

		if fileErr == nil {
			log.SetOutput(logFile)
			log.SetTimeFormat(time.Kitchen)
			log.SetReportCaller(true)
			log.SetLevel(log.DebugLevel)

			log.SetOutput(logFile)
			log.SetColorProfile(termenv.TrueColor)
			wd, err := os.Getwd()
			if err != nil {
				fmt.Println("Error getting current working dir", err)
				os.Exit(1)
			}
			log.Debug("🚀 Starting diffnav", "logFile", wd+string(os.PathSeparator)+logFile.Name())
		}
	}

	reader := bufio.NewReader(os.Stdin)
	var b strings.Builder

	for {
		r, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		_, err = b.WriteRune(r)
		if err != nil {
			fmt.Println("Error getting input:", err)
			os.Exit(1)
		}
	}

	input := ansi.Strip(b.String())
	if strings.TrimSpace(input) == "" {
		fmt.Println("No input provided, exiting")
		os.Exit(0)
	}
	p := tea.NewProgram(ui.New(input, cmd.Bool("unified")), tea.WithMouseAllMotion())

	_, err = p.Run()
	return err
}
