package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/bitzesty/ipstash/config"
	"github.com/bitzesty/ipstash/log"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
)

var dryRun bool
var ipFetchURL string
var ipStashChannel string
var rdb *redis.Client
var ip string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ipstash",
	Short: "A way to find your IP address and send it to a redis channel",
	Long: `Store your IP address in a redis channel, to later be consued for
example to update an AWS security group.
Built by Bit Zesty, for fly.io apps where the IP address changes frequently.`,
	Run: func(cmd *cobra.Command, args []string) {

		ip, err := fetchIP()

		if err != nil {
			log.Fatalf("Error fetching IP: %v", err)
			return
		}

		// If dry-run is set, just log the IP and return
		if dryRun {
			log.Infof("Dry run: IP address %s would be stored in Redis.", ip)
			return
		}

		publishIP(ip)

	},
}

var testIPCmd = &cobra.Command{
	Use:   "testip",
	Short: "Test with a user-provided IP",
	Long:  `This command allows you to test with a user-provided IP address.`,
	Run: func(cmd *cobra.Command, args []string) {
		publishIP(ip)
	},
}

func publishIP(ip string) error {
	ctx := context.Background()

	result := rdb.Publish(ctx, ipStashChannel, ip)
	if err := result.Err(); err != nil {
		log.Fatalf("Failed to publish IP to Redis channel 'ipstash': %v", err)
	} else {
		log.Infof("IP address %s published to 'ipstash' channel", ip)
	}

	return nil
}

func fetchIP() (string, error) {
	resp, err := http.Get(ipFetchURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	parsedIP := net.ParseIP(strings.TrimSpace(string(ip)))
	if parsedIP == nil {
		return "", errors.New("invalid IP format received")
	}

	return strings.TrimSpace(string(ip)), nil
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
	testIPCmd.Flags().StringVarP(&ip, "ip", "i", "", "IP address to test")
	rootCmd.AddCommand(testIPCmd)
}

func initConfig() {
	ipFetchURL = config.Config().GetString("IP_FETCH_URL")

	ipStashChannel = config.Config().GetString("IPSTASH_CHANNEL")

	redisUrl := config.Config().GetString("REDIS_URL")

	opts, err := redis.ParseURL(redisUrl)
	if err != nil {
		panic(err)
	}

	rdb = redis.NewClient(opts)
}
