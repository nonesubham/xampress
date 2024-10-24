/*
Copyright © 2024 Subham nullsubham@gmail.com
*/
package cmd

import (
	"fmt"
	"xampress/helpers"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get config data from config file.",
	Long: `Retrieves and displays configuration details from config.yaml,
including XAMPP directories and default WordPress credentials.
Use this command to view the default database credentials and 
installation details stored in the configuration file.
`,
	Run: func(cmd *cobra.Command, args []string) {
			val, err := helpers.GetAll()
			helpers.ErrHandler(err, true, "", false)
			for key, value := range val {
				fmt.Println(key, ":-")
				if nestedMap, ok := value.(map[string]interface{}); ok {
					for nestedKey, nestedValue := range nestedMap {
						if nestedValue == "" {
							nestedValue = "blank"
						}
						fmt.Printf("  %s: %v\n", nestedKey, nestedValue)
					}
				}
			}
	},
}

func init() {
	configCmd.AddCommand(getCmd)
}
