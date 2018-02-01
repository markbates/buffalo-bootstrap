package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/structs"
	"github.com/gobuffalo/buffalo/generators"
	"github.com/gobuffalo/makr"
	"github.com/gobuffalo/packr"
	"github.com/markbates/inflect"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var layout = Layout{}

// layoutCmd represents the layout command
var layoutCmd = &cobra.Command{
	Use:   "layout",
	Short: "generates a new new Bootstrap layout",
	RunE: func(cmd *cobra.Command, args []string) error {
		return layout.Run()
	},
}

func init() {
	pwd, _ := os.Getwd()
	pwd = inflect.Titleize(filepath.Base(pwd))
	layoutCmd.Flags().StringVarP(&layout.AppName, "app-name", "a", pwd, "the name of the application")
	layoutCmd.Flags().StringVarP(&layout.NavStyle, "nav-style", "n", "inverse", "style of the nav bar [default, inverse]")
	RootCmd.AddCommand(layoutCmd)
}

type Layout struct {
	AppName  string
	NavStyle string
}

func (l Layout) Run() error {
	box := packr.NewBox("./layout")
	g := makr.New()
	err := box.Walk(func(p string, f packr.File) error {
		info, err := f.FileInfo()
		if err != nil {
			return errors.WithStack(err)
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(p, ".tmpl") {
			fp := strings.TrimSuffix(p, ".tmpl")
			g.Add(makr.NewFile(fp, box.String(p)))
		}
		return nil
	})
	if err != nil {
		return errors.WithStack(err)
	}
	g.Add(&makr.Func{
		Should: func(data makr.Data) bool { return true },
		Runner: func(root string, data makr.Data) error {
			err := generators.AddInsideAppBlock(
				`app.Use(func (next buffalo.Handler) buffalo.Handler {
					return func(c buffalo.Context) error {
						c.Set("year", time.Now().Year())
						return next(c)
					}
				})`,
			)
			if err != nil {
				return errors.WithStack(err)
			}
			return generators.AddImport(filepath.Join("actions", "app.go"), "time")
		},
	})
	return g.Run(".", structs.Map(l))
}
