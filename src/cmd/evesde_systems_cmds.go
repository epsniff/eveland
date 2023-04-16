package cmd

import (
	"fmt"

	"github.com/epsniff/eveland/src/evesdedb"
	"github.com/epsniff/eveland/src/evesdk"
	"github.com/spf13/cobra"
)

// Add system commands to the root command.
func addSystemCommands(rootCmd *cobra.Command, eveSDK *evesdk.EveLand, dbpath string) {

	var systemName = "Odebeinn" // default system name
	var GetSystemIDFromNameCmd = &cobra.Command{
		Use:   "system-id",
		Short: "system-id",
		Long: `
	given a system name, it returns the system id.  Which is used by the other commands and the eve api.
	  go run main.go system-id -s=Odebeinn
	  go run main.go system-id -s=Scheenins
	`,
		Run: func(cmd *cobra.Command, args []string) {
			evesde, err := evesdedb.New(dbpath)
			if err != nil {
				fmt.Println("error: ", err)
				return
			}

			// TODO - make this a flag and lookup the region id from the db.
			if sysId, err := evesde.GetSystemID(systemName); err != nil {
				fmt.Println("error: ", err)
			} else {
				fmt.Printf("SystemName: %v System ID: %v\n", systemName, sysId)
			}
		},
	}
	GetSystemIDFromNameCmd.PersistentFlags().
		StringVarP(&systemName, "system", "s", "Odebeinn", "system name to use as the center of the search. default is Odebeinn.")

	var numberOfJumps = 5 // default number of jumps
	var JumpDistanceCmd = &cobra.Command{
		Use:   "jumps",
		Short: "jumps",
		Long: `
		given a system name to use as the center of the search, it returns all systems within N jumps.
		  go run main.go jumps -s=Odebeinn -n=5
		  go run main.go jumps -s=Scheenins -n=8
		`,
		Run: func(cmd *cobra.Command, args []string) {
			evesde, err := evesdedb.New(dbpath)
			if err != nil {
				fmt.Println("error: ", err)
				return
			}

			if sysId, err := evesde.GetSystemID(systemName); err != nil {
				fmt.Println("error: ", err)
			} else if res, err := evesde.SystemsWithinNJumps(sysId, numberOfJumps); err != nil {
				fmt.Println("error: ", err)
			} else {
				fmt.Printf("Systems within %v jumps of %v:  results:\n%v\n", 5, systemName, res)
			}
		},
	}
	JumpDistanceCmd.PersistentFlags().
		IntVarP(&numberOfJumps, "jumps", "j", 5, "number of jumps to search. default is 5.")
	JumpDistanceCmd.PersistentFlags().
		StringVarP(&systemName, "system", "s", "Odebeinn", "system name to use as the center of the search. default is Odebeinn.")

	rootCmd.AddCommand(JumpDistanceCmd)
	rootCmd.AddCommand(GetSystemIDFromNameCmd)
}
