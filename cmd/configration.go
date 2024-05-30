package cmd

import (
	"os"
	"xampress/utils"
    "encoding/json"
	"github.com/spf13/cobra"
)

var (
    sql_user string
    sql_pass string
    wp_user string
    wp_pass string
    wp_email string
)

var configCMD = &cobra.Command{
    Use:   "config",
    Short: "Configure MySQL & WordPress credentials.",
    Long:  "Configure or set new MySQL & WordPress credentials. (if default credentials are not working!)",
    Run: func(cmd *cobra.Command, args []string) {
        sqlUser, _ := cmd.Flags().GetString("sql-user")
        sqlPass, _ := cmd.Flags().GetString("sql-pass")
        wpUser, _ := cmd.Flags().GetString("wp-user")
        wpPass, _ := cmd.Flags().GetString("wp-pass")
        wpEmail, _ := cmd.Flags().GetString("wp-email")
        var cong_flags = [4]string{sqlUser,wpUser,wpPass,wpEmail}
        for i := 0; i < len(cong_flags); i++ {
            if cong_flags[i] == ""{
                utils.PrintScrn("Flags can't be empty!","red",0,false)
                os.Exit(0)
            }
        }
        newConf := utils.Config{Db_user: sqlUser, Db_pass: sqlPass, Wp_user: wpUser, Wp_pass: wpPass, Wp_email: wpEmail}
        enc_data, _ := json.Marshal(newConf)
        err := os.WriteFile("xampress/config.json", enc_data, 0644)
        utils.Chk_error(err, "Error while updating 'config.json'")
        utils.PrintScrn("Credentials updated successfully in 'config.json'","green",0,false)
    },
}


