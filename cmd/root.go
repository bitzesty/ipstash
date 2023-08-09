package cmd

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
	"time"
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
	Short: "A way to find your IP address and store it in redis",
	Long: `A simple way to store your IP address in redis.
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
		log.Debugf("%s", ip)


		// If dry-run is set, just log the IP and return
		if dryRun {
			log.Infof("Dry run: IP address %s would be stored in Redis.", ip)
			return
		}

		// Add the IP to Redis
		ctx := context.Background()
		rdb.ZAdd(ctx, "ip_addresses", &redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: ip,
		})

		// Check the number of IP addresses in Redis and remove the oldest if count > 60
		count := rdb.ZCard(ctx, "ip_addresses").Val()
		if count > 60 {
			rdb.ZRemRangeByRank(ctx, "ip_addresses", 0, 0)
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
	redisAddr := config.Config().GetString("REDIS_ADDR")

	// Provide a default value if not set
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// Initialize Redis client with fetched address
	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})
}
