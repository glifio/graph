package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// The name of our config file, without the file extension because viper supports many different config file languages.
	defaultConfigFilename = "graph"

	// The environment variable prefix of all environment variables bound to our command line flags.
	// For example, --number is bound to GIRAPH_NUMBER.
	envPrefix = "GIRAPH"
)

var (
	// Used for flags.
	cfgFile     string
	userLicense string
	
	flag_address string // blob key
	flag_method uint64 // blob raw data (or '-' for stdin)
	actorType string // blob file location

	author string

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
	fmt.Println("execute")
	return rootCmd.Execute()
}

func init() {
	fmt.Println("init")

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP("port", "p", "9111", "port number (default is 9111)")
	// rootCmd.PersistentFlags().String("db-host", "", "port number (default is 9111)")
	// rootCmd.PersistentFlags().String("db-port", "", "port number (default is 9111)")
	// rootCmd.PersistentFlags().String("db-user", "", "port number (default is 9111)")
	// rootCmd.PersistentFlags().String("db-password", "", "port number (default is 9111)")
	// rootCmd.PersistentFlags().String("db-database", "", "port number (default is 9111)")
	// rootCmd.PersistentFlags().String("lotus-token", "", "port number (default is 9111)")
	// rootCmd.PersistentFlags().String("lotus-address", "", "port number (default is 9111)")
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	//viper.BindPFlag("db-host", rootCmd.PersistentFlags().Lookup("db-host"))
	viper.BindEnv("db_host")
	viper.BindEnv("db_port")
	viper.BindEnv("db_user")
	viper.BindEnv("db_password")
	viper.BindEnv("db_database")
	viper.BindEnv("lotus_token")
	viper.BindEnv("lotus_addr")

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
	fmt.Println("initConfig")
	
	if cfgFile != "" {
		// Use config file from the flag.
		fmt.Println("config file from the flag")
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("~/.graph")
		viper.SetConfigName(defaultConfigFilename)
		viper.SetConfigType("env")
	}

	viper.SetEnvPrefix("giraph")
	viper.AutomaticEnv()
	
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}else{
		fmt.Println(err)
	}

	for _, key := range viper.AllKeys(){
		fmt.Println(key + " : " + viper.GetString(key))
	}  
}