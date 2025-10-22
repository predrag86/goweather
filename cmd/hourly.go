package cmd

import (
	"fmt"

	"goweather/internal/api"
	"goweather/internal/cache"
	"goweather/internal/cli"
	"goweather/internal/config"
	"goweather/internal/log"
	"goweather/internal/ui"

	"github.com/spf13/cobra"
)

var hoursFlag int

func init() {
	cmd := &cobra.Command{
		Use:   "hourly",
		Short: "Display hourly forecast for a city",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.Load()
			log.Init(verboseFlag)
			defer log.Sync()

			theme := ui.GetTheme(colorFlag, map[bool]string{true: "on", false: "off"}[emojiFlag])
			c := cache.NewCache(cfg.CacheDuration)
			coords, err := api.GetCoordinates(cityFlag)
			if err != nil {
				log.Logger.Fatalw("Geocoding failed", "error", err)
			}
			result, err := api.GetHourly(coords.Latitude, coords.Longitude)
			if err != nil {
				log.Logger.Fatalw("Fetch failed", "error", err)
			}
			c.Set(fmt.Sprintf("%s_hourly", cityFlag), result)
			c.BackgroundRefresh(fmt.Sprintf("%s_hourly", cityFlag), func() (any, error) {
				return api.GetHourly(coords.Latitude, coords.Longitude)
			})
			cli.PrintHourly(result, theme, hoursFlag, cfg)

		},
	}

	cmd.Flags().StringVarP(&cityFlag, "city", "c", "belgrade", "City name")
	cmd.Flags().IntVar(&hoursFlag, "hours", 6, "Number of hours to display")
	cmd.Flags().StringVar(&colorFlag, "color", "auto", "Color theme")
	cmd.Flags().BoolVar(&emojiFlag, "emoji", true, "Enable emoji output")
	cmd.Flags().BoolVar(&verboseFlag, "verbose", false, "Verbose logging")

	rootCmd.AddCommand(cmd)
}
