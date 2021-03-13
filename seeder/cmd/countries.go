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
	countriesURL = "https://download.geonames.org/export/dump/countryInfo.txt"
	filter       string

	countriesCmd = &cobra.Command{
		Use:   "countries",
		Short: "Import countries into the database",
		RunE: func(cmd *cobra.Command, args []string) error {

			// load file into memory
			resp, err := http.Get(countriesURL)
			if err != nil {
				return errors.New("Failed to download the list of countries.")
			}
			defer resp.Body.Close()

			r := csv.NewReader(resp.Body)
			r.Comma = '\t'
			r.Comment = '#'

			countries, err := r.ReadAll()
			if err != nil {
				return errors.New("Failed to parse country list.")
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

			for i, country := range countries {

				if filter != "" && filter != country[0] {
					continue
				}

				_, err = conn.Exec(context.Background(), "INSERT INTO app.countries (iso_alpha2, iso_alpha3, iso_numeric, name) VALUES ($1, $2, $3, $4)",
					country[0],
					country[1],
					country[2],
					country[4],
				)
				if err != nil {
					return fmt.Errorf("Seeding the DB failed on line %d", i+1)
				}

				numImported++
			}

			fmt.Printf("Imported %d of a possible %d countries.\n", numImported, len(countries))

			return nil
		},
	}
)

func init() {

	rootCmd.AddCommand(countriesCmd)

	countriesCmd.Flags().StringVar(&filter, "filter", "", "Filter by ISO-alpha2")
}
