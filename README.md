# compose-vault

`compose-vault` is a small CLI made for Docker Compose deployments. It parses a Compose file, finds placeholders such as `vault:secret/data/databases/postgres#username`, fetches the referenced secrets from HashiCorp Vault, and prints the rendered Compose YAML to standard output.

This tool exists for small and straightforward deployments where Docker Compose is enough and you do not want to introduce more moving parts just to inject secrets.

```yaml
# compose.yaml
services:
  db:
    image: postgres:17
    environment:
      POSTGRES_USER: vault:secret/data/databases/postgres#username
      POSTGRES_PASSWORD: vault:secret/data/databases/postgres#password
```

```bash
compose-vault render compose.yaml | docker compose -p postgres-db -f - up -d
```

## Help

```text
Usage: compose-vault <command> [flags]

Render Docker Compose files with secrets fetched from HashiCorp Vault.

Flags:
  -h, --help                Show context-sensitive help.
      --log-level="info"
      --version             Print version and quit

Commands:
  render    Render a Compose file by replacing Vault placeholders with secret values.
  check     Check YAML and Vault placeholder syntax without contacting Vault.

Run "compose-vault <command> --help" for more information on a command.
```

## Commands

### `check`

Checks YAML and placeholder syntax, it does not contact Vault.

```bash
compose-vault check compose.yaml
```

### `render`

Reads placeholders from YAML, fetches the matching values from Vault, replaces them in memory, and writes the rendered YAML to standard output.

Use it right before `docker compose up`.

```bash
compose-vault render compose.yaml
```
