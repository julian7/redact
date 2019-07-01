package main

import (
	"os"
	"strings"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var accessListCmd = &cobra.Command{
	Use:   "list",
	Short: "List collaborators to secrets in git repo",
	Run:   accessListDo,
}

func init() {
	accessCmd.AddCommand(accessListCmd)
}

func accessListDo(cmd *cobra.Command, args []string) {
	masterkey, err := basicDo()
	if err != nil {
		cmdErrHandler(err)
	}
	toplevel, err := gitutil.TopLevel()
	if err != nil {
		cmdErrHandler(err)
	}
	kxdir := files.ExchangeDir(toplevel)
	afero.Walk(masterkey.Fs, kxdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !strings.HasSuffix(path, files.ExtKeyArmor) {
			return nil
		}
		entities, err := gpgutil.LoadPubKeyFromFile(path, true)
		if err != nil {
			log.Log().Warnf("cannot load public key: %v", err)
			return nil
		}
		if len(entities) != 1 {
			log.Log().Warnf("multiple entities in key file %s", path)
			return nil
		}
		gpgutil.PrintKey(entities[0])
		return nil
	})
}
