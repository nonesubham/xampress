package cmd

import(
	"fmt"
	"github.com/spf13/cobra"
)

var cloneWP = &cobra.Command{
	Use:   "clone [projectname]",
	Short: "Clone A Exsiting WordPress Site",
	Long:  `Clone Exsiting WordPress Site From XAMPP.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Command currently not available, I am working on it :)")
	},
}

func init() {
	rootCmd.AddCommand(cloneWP)
}