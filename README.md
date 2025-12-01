# dnscli

A command-line tool for managing DNS records on Alibaba Cloud DNS, similar to the Python `autodns` module.

## Installation

### Build from source

```bash
cd dnscli
go mod tidy
go build -o dnscli main.go
```

### Install globally

```bash
go install github.com/alauda/dnscli@latest
```

## Usage

### Global Flags

**Required for all commands:**
- `--access-key-id`: Alibaba Cloud Access Key ID (required)
- `--access-key-secret`: Alibaba Cloud Access Key Secret (required)

**Required for add/delete/add-ares/delete-ares commands:**
- `--ip`: IP address to map to DNS records

**Optional:**
- `--base-domain`: Base domain name (default: "alaudatech.net")

### Commands

#### 1. Add DNS Records (Manual Mode)

Add one or more custom domain prefixes:

```bash
./dnscli add \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET" \
  --ip="123.45.67.89" \
  --domains="test-gitlab,test-jenkins"
```

This creates:
- `test-gitlab.alaudatech.net` → `123.45.67.89`
- `test-jenkins.alaudatech.net` → `123.45.67.89`

**Output:**
```
Successfully added DNS record: test-gitlab.alaudatech.net -> 123.45.67.89
Successfully added DNS record: test-jenkins.alaudatech.net -> 123.45.67.89
```

#### 2. Delete DNS Records (Manual Mode)

Delete one or more domain prefixes:

```bash
./dnscli delete \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET" \
  --ip="123.45.67.89" \
  --domains="test-gitlab,test-jenkins"
```

**Output:**
```
Successfully deleted DNS record: test-gitlab.alaudatech.net (ID: 123456789)
Successfully deleted DNS record: test-jenkins.alaudatech.net (ID: 123456790)
```

#### 3. Add DNS Records (Ares Integration Mode)

Automatically generate and add DNS records for all DevOps tools:

```bash
./dnscli add-ares \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET" \
  --ip="123.45.67.89"
```

This creates 6 DNS records:
- `123-45-67-89-jenkins.alaudatech.net` → `123.45.67.89`
- `123-45-67-89-gitlab.alaudatech.net` → `123.45.67.89`
- `123-45-67-89-sonar.alaudatech.net` → `123.45.67.89`
- `123-45-67-89-harbor.alaudatech.net` → `123.45.67.89`
- `123-45-67-89-katanomi.alaudatech.net` → `123.45.67.89`
- `123-45-67-89-nexus.alaudatech.net` → `123.45.67.89`

**Output:**
```
Adding 6 DNS records for Ares integration...
Successfully added DNS record: 123-45-67-89-jenkins.alaudatech.net -> 123.45.67.89
Successfully added DNS record: 123-45-67-89-gitlab.alaudatech.net -> 123.45.67.89
...
```

#### 4. Delete DNS Records (Ares Integration Mode)

Delete all auto-generated DevOps tool DNS records:

```bash
./dnscli delete-ares \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET" \
  --ip="123.45.67.89"
```

**Output:**
```
Deleting 6 DNS records for Ares integration...
Successfully deleted DNS record: 123-45-67-89-jenkins.alaudatech.net (ID: 123456789)
...
```

#### 5. List All DNS Records

List all DNS records under the base domain:

```bash
./dnscli list \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET"
```

**Output:**
```
DNS Records under alaudatech.net:
Domain                                   Type       Value                Status
------------------------------------------------------------------------------------
test-gitlab.alaudatech.net              A          123.45.67.89         ENABLE
test-jenkins.alaudatech.net             A          123.45.67.89         ENABLE
app1.alaudatech.net                     AAAA       2001:db8::1          ENABLE

Total: 3 record(s)
```

#### 6. Cleanup DNS Records

Cleanup (delete) specific DNS records:

```bash
./dnscli cleanup \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET" \
  --domains="test-gitlab,test-jenkins"
```

**Output:**
```
Cleaning up 2 specified domain(s)...
Deleted: test-gitlab.alaudatech.net (ID: 123456789)
Deleted: test-jenkins.alaudatech.net (ID: 123456790)
```

**Note:** You must specify which domains to cleanup using the `--domains` flag. This is a safety measure to prevent accidental deletion of all records.

### Environment Variables

You can also set credentials via environment variables to avoid passing them on command line:

```bash
export DNSCLI_ACCESS_KEY_ID="YOUR_ACCESS_KEY_ID"
export DNSCLI_ACCESS_KEY_SECRET="YOUR_ACCESS_KEY_SECRET"

# Then use commands without credential flags
./dnscli add --ip="123.45.67.89" --domains="test-gitlab"
```

To enable this, modify `cmd/root.go` to read from environment variables as fallback.

## How to Verify DNS Records

After adding DNS records, you can verify they were created successfully:

### Method 1: Use dnscli list command (Recommended)

```bash
./dnscli list \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET"
```

This will show all DNS records under your base domain.

### Method 2: Use nslookup

```bash
# Query specific domain
nslookup test-gitlab.alaudatech.net

# Use Alibaba Cloud DNS server
nslookup test-gitlab.alaudatech.net 223.5.5.5
```

### Method 3: Use dig

```bash
# Query A record
dig test-gitlab.alaudatech.net

# Query AAAA record (IPv6)
dig test-gitlab.alaudatech.net AAAA

# Use specific DNS server
dig @223.5.5.5 test-gitlab.alaudatech.net
```

### Method 4: Use host

```bash
host test-gitlab.alaudatech.net
```

### Method 5: Use ping

```bash
ping test-gitlab.alaudatech.net
```

**Note:** DNS records are added immediately in Alibaba Cloud, but global DNS cache updates may take a few minutes to several hours depending on TTL settings.

## Architecture

### Project Structure

```
dnscli/
├── main.go              # Entry point
├── cmd/
│   └── root.go          # CLI commands using cobra
└── pkg/
    └── dns/
        └── client.go    # DNS client implementation
```

### Key Components

- **DNS Client** (`pkg/dns/client.go`): Wraps Alibaba Cloud DNS SDK
  - Supports A and AAAA record types
  - Checks for existing records before adding
  - Finds and deletes records by domain prefix
  - Lists all records under base domain
  - Cleanup specific records with safety checks

- **CLI Commands** (`cmd/root.go`): Six main commands
  - `add`: Add custom domain prefixes
  - `delete`: Delete custom domain prefixes
  - `add-ares`: Auto-generate and add tool domains
  - `delete-ares`: Auto-generate and delete tool domains
  - `list`: List all DNS records
  - `cleanup`: Cleanup specific domains (requires --domains flag)

## Examples

### Single domain with IPv4

```bash
./dnscli add \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET" \
  --ip="192.168.1.100" \
  --domains="my-app"
```

### Multiple domains with IPv6

```bash
./dnscli add \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET" \
  --ip="2001:db8::1" \
  --domains="app1,app2,app3"
```

### Ares integration with custom base domain

```bash
./dnscli add-ares \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET" \
  --ip="10.0.0.50" \
  --base-domain="example.com"
```

### List and verify records

```bash
# List all records
./dnscli list \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET"

# Filter specific domain
./dnscli list \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET" | grep test-gitlab
```

### Cleanup specific domains

```bash
./dnscli cleanup \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET" \
  --domains="test-gitlab,test-jenkins"
```
## Development

### Run tests

```bash
go test ./...
```

### Build for multiple platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o dnscli-linux main.go

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o dnscli-macos-amd64 main.go

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o dnscli-macos-arm64 main.go
```

## License

MIT
