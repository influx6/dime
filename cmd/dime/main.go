package main


import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/influx6/faux/metrics"
	"github.com/influx6/faux/metrics/sentries/stdout"
	"github.com/influx6/moz/annotations"
	"github.com/influx6/moz/ast"
	"github.com/influx6/moz/gen"
	cli "gopkg.in/urfave/cli.v2"
)

var (
	version     = "0.0.1"
	commands    = []*cli.Command{}

	gupath = "github.com/gu-io/gu"
)

func main() {
	initCommands()

	app := &cli.App{}
	app.Name = "dime"
	app.Version = version
	app.Commands = commands
	app.Usage = `CLI tooling to generate service code for giving type.`

	app.Run(os.Args)
}

func capitalize(val string) string {
	return strings.ToUpper(val[:1]) + val[1:]
}

func initCommands() {

	commands = append(commands, &cli.Command{
		Name:        "generate",
		Usage:       "dime generate",
		Description: "Command will generate all needed Services and Adapters source files for annotated types.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "inputdir",
				Aliases: []string{"dir"},
				Usage:   "dir=./",
			},
		},
		Action: func(ctx *cli.Context) error {
			indir := ctx.String("inputdir")

			if indir == "" {
				cdir, err := os.Getwd()
				if err != nil {
					return err
				}

				indir = cdir
			}

			register := ast.NewAnnotationRegistry()

			generators.RegisterGenerators(register)

			// Register @assets annotation for our registery as well.
			// register.Register("assets", annotations.AssetsAnnotationGenerator)

			events := metrics.New(stdout.Stderr{})
			pkgs, err := ast.ParseAnnotations(events, indir)
			if err != nil {
				events.Emit(stdout.Error(err).With("dir", indir).With("message", "Failed to parse package annotations"))
				return err
			}

			if err := ast.Parse(events, register, pkgs...); err != nil {
				events.Emit(stdout.Error(err).With("dir", indir).With("message", "Failed to parse package annotations"))
				return err
			}

			return nil
		},
	})

	// commands = append(commands, &cli.Command{
	// 	Name:        "app",
	// 	Usage:       "gu app <package-name>",
	// 	Description: `Generates a new boilerplate for app package.`,
	// 	Flags:       []cli.Flag{},
	// 	Action: func(ctx *cli.Context) error {
	// 		args := ctx.Args()
  //
	// 		if args.Len() == 0 {
	// 			return errors.New("Please provide the name for your package")
	// 		}
  //
	// 		component := args.First()
	// 		currentDir, err := os.Getwd()
	// 		if err != nil {
	// 			return err
	// 		}
  //
	// 		directives, err := generators.GuPackageGenerator(ast.AnnotationDeclaration{Arguments: []string{component}}, ast.PackageDeclaration{FilePath: currentDir})
	// 		if err != nil {
	// 			return err
	// 		}
  //
	// 		// appDir := filepath.Join(currentDir, component)
  //
	// 		for _, directive := range directives {
	// 			if directive.Dir != "" {
	// 				coDir := filepath.Join(currentDir, directive.Dir)
  //
	// 				if _, err := os.Stat(coDir); err != nil {
	// 					drel, _ := filepath.Rel(currentDir, coDir)
	// 					fmt.Printf("- Creating package directory: %q\n", drel)
  //
	// 					if err := os.MkdirAll(coDir, 0700); err != nil && err != os.ErrExist {
	// 						return err
	// 					}
	// 				}
  //
	// 			}
  //
	// 			if directive.Writer == nil {
	// 				continue
	// 			}
  //
	// 			coFile := filepath.Join(currentDir, directive.Dir, directive.FileName)
  //
	// 			if _, err := os.Stat(coFile); err == nil {
	// 				if directive.DontOverride {
	// 					continue
	// 				}
	// 			}
  //
	// 			dFile, err := os.Create(coFile)
	// 			if err != nil {
	// 				return err
	// 			}
  //
	// 			if _, err := directive.Writer.WriteTo(dFile); err != nil {
	// 				return err
	// 			}
  //
	// 			rel, _ := filepath.Rel(currentDir, coFile)
	// 			fmt.Printf("- Add file to package directory: %q\n", rel)
  //
	// 			dFile.Close()
	// 		}
  //
	// 		return nil
	// 	},
	// })

}

func writeFile(targetFile string, data []byte) error {
	file, err := os.Create(targetFile)
	if err != nil {
		return err
	}

	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return err
	}

	return nil
}
