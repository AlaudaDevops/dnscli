package cmd

import (
	"fmt"
	"os"

	"github.com/AlaudaDevops/dnscli/pkg/dns"
	"github.com/spf13/cobra"
)

var (
	accessKeyID     string
	accessKeySecret string
	baseDomain      string
	ipAddr          string
	domainPrefixes  []string
)

var rootCmd = &cobra.Command{
	Use:   "dnscli",
	Short: "DNS management CLI for Alibaba Cloud",
	Long: `A command-line tool for managing DNS records on Alibaba Cloud DNS.
Supports adding and deleting A/AAAA records for domain prefixes.`,
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add DNS record(s)",
	Long:  `Add one or more DNS A/AAAA records to the specified domain.`,
	RunE:  runAdd,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete DNS record(s)",
	Long:  `Delete one or more DNS records from the specified domain.`,
	RunE:  runDelete,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all DNS records",
	Long:  `List all DNS records under the specified base domain.`,
	RunE:  runList,
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleanup (delete) specific DNS records",
	Long:  `Delete specific DNS records under the specified base domain. You must specify which domains to cleanup using the --domains flag.`,
	RunE:  runCleanup,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&accessKeyID, "access-key-id", "", "Alibaba Cloud Access Key ID (required)")
	rootCmd.PersistentFlags().StringVar(&accessKeySecret, "access-key-secret", "", "Alibaba Cloud Access Key Secret (required)")
	rootCmd.PersistentFlags().StringVar(&baseDomain, "base-domain", "alaudatech.net", "Base domain name")

	// IP is only required for add/delete commands
	rootCmd.PersistentFlags().StringVar(&ipAddr, "ip", "", "IP address to map")

	rootCmd.MarkPersistentFlagRequired("access-key-id")
	rootCmd.MarkPersistentFlagRequired("access-key-secret")

	// Add/Delete specific flags
	addCmd.Flags().StringSliceVar(&domainPrefixes, "domains", []string{}, "Domain prefixes to add (comma-separated)")
	addCmd.MarkFlagRequired("domains")
	addCmd.MarkFlagRequired("ip")

	deleteCmd.Flags().StringSliceVar(&domainPrefixes, "domains", []string{}, "Domain prefixes to delete (comma-separated)")
	deleteCmd.MarkFlagRequired("domains")
	deleteCmd.MarkFlagRequired("ip")

	// Cleanup specific flags
	cleanupCmd.Flags().StringSliceVar(&domainPrefixes, "domains", []string{}, "Domain prefixes to cleanup (comma-separated)")
	cleanupCmd.MarkFlagRequired("domains")

	// Add commands to root
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(cleanupCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func createClient() (*dns.Client, error) {
	cfg := &dns.Config{
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		BaseDomain:      baseDomain,
	}
	return dns.NewClient(cfg)
}

func runAdd(cmd *cobra.Command, args []string) error {
	client, err := createClient()
	if err != nil {
		return err
	}

	for _, prefix := range domainPrefixes {
		if err := client.AddDomainRecord(prefix, ipAddr); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add %s: %v\n", prefix, err)
		}
	}
	return nil
}

func runDelete(cmd *cobra.Command, args []string) error {
	client, err := createClient()
	if err != nil {
		return err
	}

	for _, prefix := range domainPrefixes {
		if err := client.DeleteDomainRecord(prefix, ipAddr); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete %s: %v\n", prefix, err)
		}
	}
	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	client, err := createClient()
	if err != nil {
		return err
	}

	records, err := client.ListAllDomainRecords()
	if err != nil {
		return err
	}

	if len(records) == 0 {
		fmt.Printf("No DNS records found under %s\n", baseDomain)
		return nil
	}

	fmt.Printf("DNS Records under %s:\n", baseDomain)
	fmt.Printf("%-40s %-10s %-20s %s\n", "Domain", "Type", "Value", "Status")
	fmt.Println("------------------------------------------------------------------------------------")
	for _, record := range records {
		fullDomain := fmt.Sprintf("%s.%s", record.RR, baseDomain)
		fmt.Printf("%-40s %-10s %-20s %s\n", fullDomain, record.Type, record.Value, record.Status)
	}
	fmt.Printf("\nTotal: %d record(s)\n", len(records))

	return nil
}

func runCleanup(cmd *cobra.Command, args []string) error {
	client, err := createClient()
	if err != nil {
		return err
	}

	return client.CleanupDomainRecords(domainPrefixes)
}
