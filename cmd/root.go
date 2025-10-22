package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "goweather",
	Short: "Command-line weather client powered by Open-Meteo",
	Long: `goweather retrieves and caches current and hourly forecasts
using the Open-Meteo public API. Example:

  goweather current --city belgrade
  goweather hourly --city belgrade --hours 6`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
