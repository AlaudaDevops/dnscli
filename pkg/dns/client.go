package dns

import (
	"fmt"
	"net"
	"os"

	alidns "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

// Client wraps Alibaba Cloud DNS operations
type Client struct {
	client     *alidns.Client
	baseDomain string
}

// Config holds the configuration for DNS client
type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	BaseDomain      string
	Endpoint        string
}

// NewClient creates a new DNS client
func NewClient(cfg *Config) (*Client, error) {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "alidns.cn-hangzhou.aliyuncs.com"
	}
	if cfg.BaseDomain == "" {
		cfg.BaseDomain = "alaudatech.net"
	}

	config := &openapi.Config{
		AccessKeyId:     tea.String(cfg.AccessKeyID),
		AccessKeySecret: tea.String(cfg.AccessKeySecret),
		Endpoint:        tea.String(cfg.Endpoint),
	}

	client, err := alidns.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create DNS client: %w", err)
	}

	return &Client{
		client:     client,
		baseDomain: cfg.BaseDomain,
	}, nil
}

// AddDomainRecord adds a DNS A or AAAA record
func (c *Client) AddDomainRecord(domainPrefix, ipAddr string) error {
	fullDomain := fmt.Sprintf("%s.%s", domainPrefix, c.baseDomain)

	// Determine record type based on IP version
	recordType := "A"
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", ipAddr)
	}
	if ip.To4() == nil {
		recordType = "AAAA"
	}

	// Check if record already exists
	exists, err := c.checkRecordExists(domainPrefix, ipAddr, recordType)
	if err != nil {
		return fmt.Errorf("failed to check existing record: %w", err)
	}
	if exists {
		fmt.Printf("Domain '%s' is already mapped to %s, skipping\n", fullDomain, ipAddr)
		return nil
	}

	// Add the record
	request := &alidns.AddDomainRecordRequest{
		DomainName: tea.String(c.baseDomain),
		RR:         tea.String(domainPrefix),
		Type:       tea.String(recordType),
		Value:      tea.String(ipAddr),
	}

	_, err = c.client.AddDomainRecord(request)
	if err != nil {
		return fmt.Errorf("failed to add domain record: %w", err)
	}

	fmt.Printf("Successfully added DNS record: %s -> %s\n", fullDomain, ipAddr)
	return nil
}

// DeleteDomainRecord deletes a DNS record
func (c *Client) DeleteDomainRecord(domainPrefix, ipAddr string) error {
	fullDomain := fmt.Sprintf("%s.%s", domainPrefix, c.baseDomain)

	// Find the record ID
	recordID, err := c.findRecordID(domainPrefix)
	if err != nil {
		return fmt.Errorf("failed to find record: %w", err)
	}
	if recordID == "" {
		fmt.Printf("Domain '%s' does not exist, no cleanup needed\n", fullDomain)
		return nil
	}

	// Delete the record
	request := &alidns.DeleteDomainRecordRequest{
		RecordId: tea.String(recordID),
	}

	_, err = c.client.DeleteDomainRecord(request)
	if err != nil {
		return fmt.Errorf("failed to delete domain record: %w", err)
	}

	fmt.Printf("Successfully deleted DNS record: %s (ID: %s)\n", fullDomain, recordID)
	return nil
}

// checkRecordExists checks if a DNS record already exists with the given IP
func (c *Client) checkRecordExists(domainPrefix, ipAddr, recordType string) (bool, error) {
	request := &alidns.DescribeDomainRecordsRequest{
		DomainName: tea.String(c.baseDomain),
		RRKeyWord:  tea.String(domainPrefix),
		Type:       tea.String(recordType),
	}

	response, err := c.client.DescribeDomainRecords(request)
	if err != nil {
		return false, err
	}

	if response.Body.DomainRecords == nil || response.Body.DomainRecords.Record == nil {
		return false, nil
	}

	for _, record := range response.Body.DomainRecords.Record {
		if tea.StringValue(record.RR) == domainPrefix && tea.StringValue(record.Value) == ipAddr {
			return true, nil
		}
	}

	return false, nil
}

// findRecordID finds the record ID for a given domain prefix
func (c *Client) findRecordID(domainPrefix string) (string, error) {
	request := &alidns.DescribeDomainRecordsRequest{
		DomainName: tea.String(c.baseDomain),
		RRKeyWord:  tea.String(domainPrefix),
	}

	response, err := c.client.DescribeDomainRecords(request)
	if err != nil {
		return "", err
	}

	if response.Body.DomainRecords == nil || response.Body.DomainRecords.Record == nil {
		return "", nil
	}

	if len(response.Body.DomainRecords.Record) > 0 {
		return tea.StringValue(response.Body.DomainRecords.Record[0].RecordId), nil
	}

	return "", nil
}

// ListAllDomainRecords lists all DNS records under the base domain
func (c *Client) ListAllDomainRecords() ([]*DomainRecord, error) {
	request := &alidns.DescribeDomainRecordsRequest{
		DomainName: tea.String(c.baseDomain),
		PageSize:   tea.Int64(100), // Get up to 100 records per page
	}

	response, err := c.client.DescribeDomainRecords(request)
	if err != nil {
		return nil, fmt.Errorf("failed to list domain records: %w", err)
	}

	if response.Body.DomainRecords == nil || response.Body.DomainRecords.Record == nil {
		return []*DomainRecord{}, nil
	}

	var records []*DomainRecord
	for _, record := range response.Body.DomainRecords.Record {
		records = append(records, &DomainRecord{
			ID:     tea.StringValue(record.RecordId),
			RR:     tea.StringValue(record.RR),
			Type:   tea.StringValue(record.Type),
			Value:  tea.StringValue(record.Value),
			Status: tea.StringValue(record.Status),
		})
	}

	return records, nil
}

// CleanupDomainRecords deletes specific DNS records
func (c *Client) CleanupDomainRecords(domainPrefixes []string) error {
	fmt.Printf("Cleaning up %d specified domain(s)...\n", len(domainPrefixes))
	for _, prefix := range domainPrefixes {
		// Find the record ID
		recordID, err := c.findRecordID(prefix)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to find record for %s: %v\n", prefix, err)
			continue
		}
		if recordID == "" {
			fmt.Printf("Domain '%s.%s' does not exist, skipping\n", prefix, c.baseDomain)
			continue
		}

		// Delete the record
		request := &alidns.DeleteDomainRecordRequest{
			RecordId: tea.String(recordID),
		}

		_, err = c.client.DeleteDomainRecord(request)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete %s.%s: %v\n", prefix, c.baseDomain, err)
			continue
		}
		fmt.Printf("Deleted: %s.%s (ID: %s)\n", prefix, c.baseDomain, recordID)
	}
	return nil
}

// DomainRecord represents a DNS record
type DomainRecord struct {
	ID     string
	RR     string
	Type   string
	Value  string
	Status string
}

// GenerateToolDomains generates domain names for common DevOps tools based on IP
func GenerateToolDomains(ipAddr string) []string {
	tools := []string{"jenkins", "gitlab", "sonar", "harbor", "katanomi", "nexus"}
	// Replace : and . with - in IP address
	name := ipAddr
	name = replaceAll(name, ":", "-")
	name = replaceAll(name, ".", "-")

	var domains []string
	for _, tool := range tools {
		domains = append(domains, fmt.Sprintf("%s-%s", name, tool))
	}
	return domains
}

// replaceAll is a helper function to replace all occurrences
func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if i <= len(s)-len(old) && s[i:i+len(old)] == old {
			result += new
			i += len(old) - 1
		} else {
			result += string(s[i])
		}
	}
	return result
}
