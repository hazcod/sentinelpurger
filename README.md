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
