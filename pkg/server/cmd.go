package server

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/docker/distribution/uuid"

	"github.com/openshift/oscanner/pkg/database"
	"github.com/openshift/oscanner/pkg/instance"
	"github.com/openshift/oscanner/pkg/version"
)

var showVersion bool

func init() {
	RootCmd.AddCommand(ServeCmd)

	RootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show the version and exit")
}

// RootCmd is the main command for the 'registry' binary.
var RootCmd = &cobra.Command{
	Use:   "oscanner",
	Short: "`oscanner`",
	Long:  "`oscanner`",
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Printf("%v\n", version.Get())
			return
		}
		cmd.Usage()
	},
}

// ServeCmd is a cobra command for running the registry.
var ServeCmd = &cobra.Command{
	Use:   "serve <config>",
	Short: "`serve` serves requests to scan",
	Long:  "`serve` serves requests to scan.",
	Run: func(cmd *cobra.Command, args []string) {

		config, err := resolveConfiguration(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
			cmd.Usage()
			os.Exit(1)
		}

		if err := configureLogging(config); err != nil {
			fmt.Fprintf(os.Stderr, "error configuring logger: %v", err)
			os.Exit(1)
		}

		if err := database.NewDB(config).Init(); err != nil {
			log.Fatalln(err)
		}

		// inject a logger into the uuid library. warns us if there is a problem
		// with uuid generation under low entropy.
		uuid.Loggerf = log.Warnf

		in, err := instance.NewInstance(config)
		if err != nil {
			log.Fatalln(err)
		}

		// TODO cleanup processing directory

		registry, err := NewServer(in)
		if err != nil {
			log.Fatalln(err)
		}

		if err = registry.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	},
}
