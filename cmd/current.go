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

var (
	cityFlag    string
	colorFlag   string
	emojiFlag   bool
	verboseFlag bool
)

func init() {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Display current weather for a city",
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
			result, err := api.GetWeather(coords.Latitude, coords.Longitude)
			if err != nil {
				log.Logger.Fatalw("Fetch failed", "error", err)
			}
			c.Set(fmt.Sprintf("%s_current", cityFlag), result)
			cli.PrintCurrent(result, theme)
		},
	}

	cmd.Flags().StringVarP(&cityFlag, "city", "c", "belgrade", "City name")
	cmd.Flags().StringVar(&colorFlag, "color", "auto", "Color theme: auto|dark|light|none")
	cmd.Flags().BoolVar(&emojiFlag, "emoji", true, "Enable emoji output")
	cmd.Flags().BoolVar(&verboseFlag, "verbose", false, "Verbose logging")

	rootCmd.AddCommand(cmd)
}
