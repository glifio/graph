package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// The name of our config file, without the file extension because viper supports many different config file languages.
	//defaultConfigFilename = "mainnet"
	defaultConfigFilename = "graph"

	// The environment variable prefix of all environment variables bound to our command line flags.
	// For example, --number is bound to GRAPH_NUMBER.
	envPrefix = "GRAPH"
)

var (
	// Used for flags.
	cfgFile string

	flag_address string // blob key
	flag_method  uint64 // blob raw data (or '-' for stdin)

	rootCmd = &cobra.Command{
		Use:   "graph",
		Short: "A graphQL interface to Lily",
		Long: `Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Port: " + viper.GetString("port"))
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./graph.env)")
	rootCmd.PersistentFlags().StringP("port", "p", "9090", "port number (default 9090)")
	rootCmd.PersistentFlags().String("path", "./data", "path to kv db")
	rootCmd.PersistentFlags().Uint("rpc", 9091, "rpc port number (default 9091)")

	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("path", rootCmd.PersistentFlags().Lookup("path"))
	viper.BindPFlag("rpc", rootCmd.PersistentFlags().Lookup("rpc"))

	// viper.BindEnv("port")
	// viper.BindEnv("db_host")
	// viper.BindEnv("db_port")
	// viper.BindEnv("db_user")
	// viper.BindEnv("db_password")
	// viper.BindEnv("db_database")
	// viper.BindEnv("db_schema")
	viper.BindEnv("lotus_token")
	viper.BindEnv("lily")
	viper.BindEnv("lotus")
	viper.BindEnv("lotus_wss")
	viper.BindEnv("sentry")
	viper.BindEnv("confidence")

	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			viper.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefix, envVarSuffix))
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			rootCmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		//fmt.Printf("config file from the flag %s\n", cfgFile)
		viper.AddConfigPath(".")
		viper.SetConfigName(cfgFile)
		viper.SetConfigType("env")
	} else {
		viper.AddConfigPath(".")
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
		}
		viper.SetConfigName(defaultConfigFilename)
		viper.SetConfigType("env")
	}

	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
	}

	// for _, key := range viper.AllKeys() {
	// 	fmt.Println(key + " : " + viper.GetString(key))
	// }
}
