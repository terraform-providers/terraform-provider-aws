package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/andreyvit/diff"
	c "github.com/gookit/color"
	"github.com/hashicorp/go-multierror"
	"github.com/katbyte/terrafmt/lib/blocks"
	"github.com/katbyte/terrafmt/lib/common"
	"github.com/katbyte/terrafmt/lib/format"
	"github.com/katbyte/terrafmt/lib/upgrade012"
	"github.com/katbyte/terrafmt/lib/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Make() *cobra.Command {
	root := &cobra.Command{
		Use:           "terrafmt [fmt|diff|blocks|upgrade012]",
		Short:         "terrafmt is a small utility to format terraform blocks found in files.",
		Long:          `A small utility that formats terraform blocks found in files. Primarily intended to help with terraform provider development.`,
		Args:          cobra.RangeArgs(0, 0),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("No command specified")
		},
	}

	//options : only count, blocks diff/found, total lines diff, etc
	fmtCmd := &cobra.Command{
		Use:   "fmt [path]",
		Short: "formats terraform blocks in a directory, file, or stdin",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) == 1 {
				path = args[0]
			}
			common.Log.Debugf("terrafmt  %s", path)

			pattern, _ := cmd.Flags().GetString("pattern")
			filenames, err := allFiles(path, pattern)

			if err != nil {
				return err
			}

			var errs *multierror.Error
			var exitStatus int

			for _, filename := range filenames {
				blocksFormatted := 0

				br := blocks.Reader{
					LineRead: blocks.ReaderPassthrough,
					BlockRead: func(br *blocks.Reader, i int, b string) error {
						var fb string
						var err error
						if viper.GetBool("fmtcompat") {
							fb, err = format.FmtVerbBlock(b, filename)
						} else {
							fb, err = format.Block(b, filename)
						}

						if err != nil {
							return err
						}

						_, err = br.Writer.Write([]byte(fb))

						if err == nil && fb != b {
							blocksFormatted++
						}

						return err
					},
				}
				err := br.DoTheThing(filename)

				fc := "magenta"
				if blocksFormatted > 0 {
					fc = "lightMagenta"
				}

				if viper.GetBool("verbose") {
					// nolint staticcheck
					fmt.Fprintf(os.Stderr, c.Sprintf("<%s>%s</>: <cyan>%d</> lines & formatted <yellow>%d</>/<yellow>%d</> blocks!\n", fc, br.FileName, br.LineCount, blocksFormatted, br.BlockCount))
				}

				if err != nil {
					errs = multierror.Append(errs, err)
				}

				if br.ErrorBlocks > 0 {
					exitStatus = 1
				}
			}

			if errs != nil {
				return errs
			}

			if exitStatus != 0 {
				os.Exit(exitStatus)
			}

			return nil
		},
	}

	root.AddCommand(fmtCmd)
	fmtCmd.Flags().StringP("pattern", "p", "", "glob pattern to match with each file name (e.g. *.markdown)")

	//options : only count, blocks diff/found, total lines diff, etc
	root.AddCommand(&cobra.Command{
		Use:   "upgrade012 [file]",
		Short: "formats terraform blocks to 0.12 format in a single file or on stdin",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := ""
			if len(args) == 1 {
				filename = args[0]
			}
			common.Log.Debugf("terrafmt upgrade012 %s", filename)

			blocksFormatted := 0
			br := blocks.Reader{
				LineRead: blocks.ReaderPassthrough,
				BlockRead: func(br *blocks.Reader, i int, b string) error {
					var fb string
					var err error
					if viper.GetBool("fmtcompat") {
						fb, err = upgrade012.Upgrade12VerbBlock(b)
					} else {
						fb, err = upgrade012.Block(b)
					}

					if err != nil {
						return err
					}

					if _, err = br.Writer.Write([]byte(fb)); err == nil && fb != b {
						blocksFormatted++
					}

					return nil
				},
			}
			err := br.DoTheThing(filename)
			if err != nil {
				return err
			}

			fc := "magenta"
			if blocksFormatted > 0 {
				fc = "lightMagenta"
			}

			if viper.GetBool("verbose") {
				// nolint staticcheck
				fmt.Fprintf(os.Stderr, c.Sprintf("<%s>%s</>: <cyan>%d</> lines & formatted <yellow>%d</>/<yellow>%d</> blocks!\n", fc, br.FileName, br.LineCount, blocksFormatted, br.BlockCount))
			}

			if br.ErrorBlocks > 0 {
				os.Exit(-1)
			}

			return nil
		},
	})

	//options : only count, blocks diff/found, total lines diff, etc
	diffCmd := &cobra.Command{
		Use:   "diff [path]",
		Short: "formats terraform blocks in a directory, file, or stdin and shows the difference",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) == 1 {
				path = args[0]
			}
			common.Log.Debugf("terrafmt fmt %s", path)

			pattern, _ := cmd.Flags().GetString("pattern")
			filenames, err := allFiles(path, pattern)

			if err != nil {
				return err
			}

			var errs *multierror.Error
			var exitStatus int
			var hasDiff bool

			for _, filename := range filenames {
				blocksWithDiff := 0
				br := blocks.Reader{
					ReadOnly: true,
					LineRead: blocks.ReaderPassthrough,
					BlockRead: func(br *blocks.Reader, i int, b string) error {
						var fb string
						var err error
						if viper.GetBool("fmtcompat") {
							fb, err = format.FmtVerbBlock(b, filename)
						} else {
							fb, err = format.Block(b, filename)
						}
						if err != nil {
							return err
						}

						if fb == b {
							return nil
						}
						blocksWithDiff++

						// nolint staticcheck
						fmt.Fprintf(os.Stdout, c.Sprintf("<lightMagenta>%s</><darkGray>:</><magenta>%d</>\n", br.FileName, br.LineCount-br.BlockCurrentLine))

						if !viper.GetBool("quiet") {
							d := diff.LineDiff(b, fb)
							scanner := bufio.NewScanner(strings.NewReader(d))
							for scanner.Scan() {
								l := scanner.Text()
								if strings.HasPrefix(l, "+") {
									fmt.Fprint(os.Stdout, c.Sprintf("<green>%s</>\n", l))
								} else if strings.HasPrefix(l, "-") {
									fmt.Fprint(os.Stdout, c.Sprintf("<red>%s</>\n", l))
								} else {
									fmt.Fprint(os.Stdout, l+"\n")
								}
							}
						}

						return nil
					},
				}

				err := br.DoTheThing(filename)
				if err != nil {
					errs = multierror.Append(errs, err)
					continue
				}

				if blocksWithDiff > 0 {
					hasDiff = true
				}

				fc := "magenta"
				if hasDiff {
					fc = "lightMagenta"
				}

				if viper.GetBool("verbose") {
					// nolint staticcheck
					fmt.Fprintf(os.Stderr, c.Sprintf("<%s>%s</>: <cyan>%d</> lines & <yellow>%d</>/<yellow>%d</> blocks need formatting.\n", fc, br.FileName, br.LineCount, blocksWithDiff, br.BlockCount))
				}

				if br.ErrorBlocks > 0 {
					exitStatus = 1
				}
			}

			if errs != nil {
				return errs
			}

			if viper.GetBool("check") && hasDiff {
				exitStatus = 1
			}

			if exitStatus != 0 {
				os.Exit(exitStatus)
			}

			return nil
		},
	}

	root.AddCommand(diffCmd)
	diffCmd.Flags().StringP("pattern", "p", "", "glob pattern to match with each file name (e.g. *.markdown)")

	// options
	root.AddCommand(&cobra.Command{
		Use:   "blocks [file]",
		Short: "extracts terraform blocks from a file ",
		//options: no header (######), format (json? xml? etc), only should block x?
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := ""
			if len(args) == 1 {
				filename = args[0]
			}
			common.Log.Debugf("terrafmt blocks %s", filename)

			br := blocks.Reader{
				ReadOnly: true,
				LineRead: blocks.ReaderIgnore,
				BlockRead: func(br *blocks.Reader, i int, b string) error {
					// nolint staticcheck
					fmt.Fprintf(os.Stdout, c.Sprintf("\n<white>#######</> <cyan>B%d</><darkGray> @ #%d</>\n", br.BlockCount, br.LineCount))
					fmt.Fprint(os.Stdout, b)
					return nil
				},
			}

			err := br.DoTheThing(filename)

			if err != nil {
				return err
			}

			//blocks
			// nolint staticcheck
			fmt.Fprintf(os.Stderr, c.Sprintf("\nFinished processing <cyan>%d</> lines <yellow>%d</> blocks!\n", br.LineCount, br.BlockCount))

			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number of terrafmt",
		Args:  cobra.NoArgs,
		Run:   versionCmd,
	})

	pflags := root.PersistentFlags()
	pflags.BoolP("fmtcompat", "f", false, "enable format string (%s, %d etc) compatibility")
	pflags.BoolP("check", "c", false, "return an error during diff if formatting is required")
	pflags.BoolP("verbose", "v", false, "show files as they are processed& additional stats")
	pflags.BoolP("quiet", "q", false, "quiet mode, only shows block line numbers ")

	if err := viper.BindPFlag("fmtcompat", pflags.Lookup("fmtcompat")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("check", pflags.Lookup("check")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("quiet", pflags.Lookup("quiet")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("verbose", pflags.Lookup("verbose")); err != nil {
		panic(err)
	}

	//todo bind to env?

	return root
}

func allFiles(path string, pattern string) ([]string, error) {
	if path == "" {
		return []string{""}, nil
	}

	info, err := os.Stat(path)

	if err != nil {
		return nil, fmt.Errorf("error reading path (%s): %s", path, err)
	}

	if !info.IsDir() {
		return []string{path}, nil
	}

	var filenames []string

	err = filepath.Walk(
		path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if pattern == "" {
				filenames = append(filenames, path)

				return nil
			}

			matched, err := filepath.Match(pattern, filepath.Base(path))

			if err != nil {
				return err
			}

			if matched {
				filenames = append(filenames, path)
			}

			return nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("error walking path (%s): %s", path, err)
	}

	return filenames, nil
}

func versionCmd(cmd *cobra.Command, args []string) {
	// nolint errcheck
	fmt.Println("terrafmt v" + version.Version + "-" + version.GitCommit)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	tfCmd := exec.Command("terraform", "version")
	tfCmd.Stdout = stdout
	tfCmd.Stderr = stderr
	if err := tfCmd.Run(); err != nil {
		common.Log.Warnf("Error running terraform: %s", err)
		return
	}
	terraformVersion := strings.SplitN(stdout.String(), "\n", 2)[0]
	// nolint errcheck
	fmt.Println("  + " + terraformVersion)
}
