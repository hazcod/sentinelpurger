# sentinelpurger

A Go program that is able to purge logs from Microsoft Log Analytics/Sentinel according to custom retention rules.</br>
Log Analytics only supports retention of 30 days minimum in the web UI, hence this project.

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

tables:
- 
  name: AWSVPCFlow
  retention: 336h # 14 days
```

And now run the program from source code:
```shell
% make
go run ./cmd/... -config=config.yml
```

Or binary:
```shell
% sentinelpurger -config=config.yml
```

## Building

```shell
% make build
```

## Setup

1. Create an App Registration and secret key.
2. Go to your Analytics Workspace > IAM > App Roles and assign the `Data Purger` role to the app registration you created.
3. If you have a Log Search alert rule for operational issue alerting, modify is so that sucessful purging is ignored:
```azure
_LogOperation
| where Level == "Warning"
| where Detail !contains "Purge operation started" and Detail !contains "Purge opertion completed"
```