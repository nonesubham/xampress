/*
Copyright © 2024 Subham nullsubham@gmail.com
*/
package cmd

import (
	"fmt"
	"strings"
	"xampress/helpers"
	"github.com/spf13/pflag"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "A brief description of your command",
	Long: `Update configuration details in config.yaml,
including XAMPP directories and default WordPress credentials.
Use this command to update the database credentials and 
installation details stored in the configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Flags().Visit(func(flag *pflag.Flag){
			var itm [2]string = [2]string(strings.Split(flag.Name, "-"))
			var parent string
			switch itm[0] {
			case "wp":
				parent = "wordpress" 
			case "db":
				parent = "database"
			case "dir":
				parent = "xampp"
			}
			var err error = helpers.SetConf(parent+"."+itm[1], flag.Value)
			helpers.ErrHandler(err, true, "", false)
			fmt.Println(itm[1], "of", parent, "updated successfully.")
		})
	},
}

func init() {
	configCmd.AddCommand(setCmd)
	setCmd.Flags().String("wp-user", "", "Set username for WordPress's admin for new installation.")
	setCmd.Flags().String("wp-pass", "", "Set password for WordPress's admin for new installation.")
	setCmd.Flags().String("wp-email", "", "Set email for WordPress's admin for new installation.")
	setCmd.Flags().String("db-user", "", "Set Mysql user for connection. (Change it only if default user not works)")
	setCmd.Flags().String("db-pass", "", "Set Mysql password for connection. (Change it only if default password not works)")
	setCmd.Flags().String("dir-bins", "", "Set XAMPP base dir of all bins like apache, mysql & php. (Change it only if you have installed XAMPP in custom directory)")
	setCmd.Flags().String("dir-htdocs", "", "Set XAMPP's htdocs directory. (Change it only if you have installed XAMPP in custom directory)")
	setCmd.Flags().String("dir-xampp", "", "Set XAMPP installation directory (Change it only if you have installed XAMPP in custom directory)")
}
