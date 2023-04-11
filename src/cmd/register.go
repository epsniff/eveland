package cmd

import (
	"github.com/epsniff/eveland/src/evesdk"
	"github.com/spf13/cobra"
)

// Register is a helper function to register a command with the root command.
func Register(cmd *cobra.Command, eveSDK *evesdk.EveLand, dbpath string) {
	addMarketOrdersCommands(cmd, eveSDK, dbpath)
	addRegionCommands(cmd, eveSDK, dbpath)
	addItemCommands(cmd, eveSDK, dbpath)
	addSystemCommands(cmd, eveSDK, dbpath)
	addSDEUtilsCommands(cmd, eveSDK, dbpath)

	addTradersToolsCommands(cmd, eveSDK, dbpath)
}
