package cmd

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	tzURL = "https://download.geonames.org/export/dump/timeZones.txt"

	tzCmd = &cobra.Command{
		Use:   "tz",
		Short: "Import timezones into the database",
		RunE: func(cmd *cobra.Command, args []string) error {

			// load file into memory
			resp, err := http.Get(tzURL)
			if err != nil {
				return errors.New("Failed to download the list of timezones.")
			}
			defer resp.Body.Close()

			r := csv.NewReader(resp.Body)
			r.Comma = '\t'
			r.Comment = '#'

			timezones, err := r.ReadAll()
			if err != nil {
				return errors.New("Failed to parse timezone list.")
			}

			viper.SetDefault("db_user", "app")
			viper.SetDefault("db_host", "127.0.0.1")
			viper.SetDefault("db_port", 9001)
			viper.SetDefault("db_name", "app")

			// Attempt to load config values from the `.env` file. If the file is not
			// found, that's okay.
			viper.SetConfigFile("../.env")
			viper.ReadInConfig()

			// Attempt to load config values from environment variables. Most useful in
			// non development environments.
			viper.AutomaticEnv()

			connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", viper.GetString("db_user"), viper.GetString("db_pass"), viper.GetString("db_host"), viper.GetInt("db_port"), viper.GetString("db_name"))
			conn, err := pgx.Connect(context.Background(), connectionString)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
				os.Exit(1)
			}
			defer conn.Close(context.Background())

			var numImported int

			for i, tz := range timezones {

				if filter != "" && filter != tz[0] {
					continue
				}

				_, err = conn.Exec(context.Background(), "INSERT INTO app.timezones (name, iso_alpha2, gmt_offset, dst_offset, raw_offset) VALUES ($1, $2, $3, $4, $5)",
					tz[1],
					tz[0],
					tz[2],
					tz[3],
					tz[4],
				)
				if err != nil {
					return fmt.Errorf("Seeding the DB failed on line %d. Msg: %s", i+1, err)
				}

				numImported++
			}

			fmt.Printf("Imported %d of a possible %d timezones.\n", numImported, len(timezones))

			return nil
		},
	}
)

func init() {

	rootCmd.AddCommand(tzCmd)

	tzCmd.Flags().StringVar(&filter, "filter", "", "Filter by ISO-alpha2")
}
