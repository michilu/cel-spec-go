package gen

import (
	valid "github.com/asaskevich/govalidator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"

	"github.com/michilu/boilerplate/v/errs"
)

type (
	flag struct {
		dirpath string
	}

	opt struct {
		F string `valid:"direxists"`
	}
)

// New returns a new command.
func New() (*cobra.Command, error) {
	const op = "cmd.gen.New"
	f := &flag{}
	c := &cobra.Command{
		Use:   "gen",
		Short: "gen",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRunE(cmd, args, f)
		},
		Run: func(cmd *cobra.Command, args []string) {
			run(cmd, args, f)
		},
	}
	c.Flags().StringVarP(&f.dirpath, "dir", "d", "", "path to an exists dir. default: current directory")
	err := viper.BindPFlag("dir", c.Flags().Lookup("dir"))
	if err != nil {
		return nil, &errs.Error{Op: op, Err: err}
	}
	return c, nil
}

func preRunE(cmd *cobra.Command, args []string, f *flag) error {
	const op = "cmd.gen.preRunE"
	ok, err := valid.ValidateStruct(&opt{f.dirpath})
	if err != nil {
		return &errs.Error{Op: op, Err: err}
	}
	if !ok {
		return &errs.Error{Op: op, Code: codes.InvalidArgument, Message: "invalid arguments"}
	}
	return nil
}

func run(cmd *cobra.Command, args []string, f *flag) {
	gen(f.dirpath)
}
