# dnscli

A command-line tool for managing DNS records on Alibaba Cloud DNS.

## Installation

```bash
cd dnscli
go mod tidy
go build -o dnscli main.go
```

## Usage

### Commands

#### Add DNS Records

```bash
./dnscli add \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET" \
  --ip="123.45.67.89" \
  --domains="test-gitlab,test-jenkins"
```

#### Delete DNS Records

```bash
./dnscli delete \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET" \
  --ip="123.45.67.89" \
  --domains="test-gitlab,test-jenkins"
```

#### List All DNS Records

```bash
./dnscli list \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET"
```

#### Cleanup DNS Records

```bash
./dnscli cleanup \
  --access-key-id="YOUR_ACCESS_KEY_ID" \
  --access-key-secret="YOUR_ACCESS_KEY_SECRET" \
  --domains="test-gitlab,test-jenkins"
```

## Flags

**Required:**
- `--access-key-id`: Alibaba Cloud Access Key ID
- `--access-key-secret`: Alibaba Cloud Access Key Secret

**Optional:**
- `--base-domain`: Base domain name (default: "alaudatech.net")
- `--ip`: IP address (required for add/delete commands)
- `--domains`: Domain prefixes, comma-separated

## License

MIT
