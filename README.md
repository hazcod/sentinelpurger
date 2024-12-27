# miro2sentinel

A Go program that exports Miro audit logs to Microsoft Sentinel SIEM.
Two tables are used; `MiroAuditLogs`.

## Running

First create a yaml file, such as `config.yml`:
```yaml
log:
  level: INFO

microsoft:
  app_id: ""
  secret_key: ""
  tenant_id: ""
  subscription_id: ""
  resource_group: ""
  workspace_name: ""
  retention_days: 90

  dcr:
    endpoint: ""
    rule_id: ""
    stream_name: ""

  expires_months: 6

miro:
  lookback_days: 7
  access_token: "" # non-expiring Miro access token

```

And now run the program from source code:
```shell
% make
go run ./cmd/... -config=dev.yml
INFO[0000] shipping logs                                 module=sentinel_logs table_name=MiroAuditLogs total=82
INFO[0002] shipped logs                                  module=sentinel_logs table_name=MiroAuditLogs
INFO[0002] successfully sent logs to sentinel            total=82
```

Or binary:
```shell
% miro2sen -config=config.yml
```

## Building

```shell
% make build
```
