package cmd

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
	"context"
	"github.com/bitzesty/ipstash/config"
	"github.com/bitzesty/ipstash/log"
)

// Initialize a Redis client
var rdb *redis.Client
var dryRun bool  // Variable to hold the value of the dry-run flag
// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ipstash",
	Short: "A way to find your IP address and send it to a redis channel",
	Long: `Store your IP address in a redis channel, to later be consued to
update a AWS security group.
Built by Bit Zesty, for fly.io apps where the IP address changes frequently.`,
	Run: func(cmd *cobra.Command, args []string) { 
		// Fetch IP Address
		resp, err := http.Get("http://ipinfo.io/ip")
		if err != nil {
			fmt.Println("Error fetching IP:", err)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			return
		}

		ip := strings.TrimSpace(string(body))

		// If dry-run is set, just log the IP and return
		if dryRun {
			log.Infof("Dry run: IP address %s would be stored in Redis.", ip)
			return
		}

		ctx := context.Background()

		// Try to publish the IP to the "ipstash" channel
		result := rdb.Publish(ctx, "ipstash", ip)
		if err := result.Err(); err != nil {
			log.Errorf("Failed to publish IP to Redis channel 'ipstash': %v", err)
		} else {
			log.Infof("IP address %s published to 'ipstash' channel", ip)
		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Add the dry-run flag
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Perform a dry run without storing the IP in Redis")
}

func initConfig() {
	// Fetch Redis address from the existing Viper configuration
	redisUrl := config.Config().GetString("REDIS_URL")

	opts, err := redis.ParseURL(redisUrl)
    if err != nil {
        panic(err)
    }

	// Initialize Redis client with fetched address
	rdb = redis.NewClient(opts)
}
