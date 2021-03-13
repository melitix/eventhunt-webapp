package cmd

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	citiesURL = "https://download.geonames.org/export/dump/cities15000.zip"

	citiesCmd = &cobra.Command{
		Use:   "cities",
		Short: "Import cities data into the database",
		RunE: func(cmd *cobra.Command, args []string) error {

			// load zip file into memory
			resp, err := http.Get(citiesURL)
			if err != nil {
				return errors.New("Failed to download cities list.")
			}
			defer resp.Body.Close()

			zipFile, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return errors.New("Failed to read cities zip file.")
			}

			zr, err := zip.NewReader(bytes.NewReader(zipFile), int64(len(zipFile)))
			if err != nil {
				return errors.New("Failed to open cities zip file.")
			}

			citiesFile, err := zr.File[0].Open()
			if err != nil {
				return errors.New("Failed to read cities txt file.")
			}

			r := csv.NewReader(citiesFile)
			r.Comma = '\t'
			r.Comment = '#'

			cities, err := r.ReadAll()
			if err != nil {
				return errors.New("Failed to parse cities list.")
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

			for i, city := range cities {

				if filter != "" && filter != city[8] {
					continue
				}

				id := city[0]
				iso2 := city[8]
				code := city[10]
				name := city[1]
				lat := city[4]
				long := city[5]
				timezone := city[17]

				_, err = conn.Exec(context.Background(), "INSERT INTO app.cities (id, iso_alpha2, admin1, name, location, timezone) VALUES ($1, $2, $3, $4, $5, $6)",
					id,
					iso2,
					code,
					name,
					"SRID=4326;POINT("+long+" "+lat+")",
					timezone,
				)
				if err != nil {
					return fmt.Errorf("Seeding the DB failed on line %d. Msg: %s", i+1, err)
				}

				numImported++
			}

			fmt.Printf("Imported %d of a possible %d cities.\n", numImported, len(cities))

			return nil
		},
	}
)

func init() {

	rootCmd.AddCommand(citiesCmd)

	citiesCmd.Flags().StringVar(&filter, "filter", "", "Filter by ISO-alpha2")
}
