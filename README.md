# sbercdn-exporter

Prometheus metrics exporter for SberCloud CDN https://cdn.sber.cloud.
API documentation https://docs.sbercloud.ru/cdn/ug/topics/guides__api.html

## Configuring
At this moment service may be configured from config and/or from env.
### From config file
There is only one cli flag now `-config` to get settings from YAML formatted config file for eg.:
```
listen:
  address: ":9921" # optional, address to listen on defaults to :9921
  cert_file:       # optional, path certificate file for endpoint encryption
  privkey_file:    # optional, path to unencrypted private key file for endpoint encryption
                   # if both cert_file and privkey_file are not empty exporter will serve metrics through HTTPS

api:
  url: "https://api.cdn.sber.cloud" # optional, API URL, defaults to https://api.cdn.sber.cloud
  username: "username@example.com"  # mandatory, API username
  password: "password"              # mandatory, API password
  accounts: []                      # optional, limits used accounts, by default used all found active accounts
  token_lifetime: "6h"              # optional, API token lifetime defaults to 6 hours
  max_query_time: "10s"             # optional, defaults to 10 seconds, maximum time for API request, all incomplete requests canceled when time exeedes
  scrape_time_offset: "5m"          # optional, default to 5 minutes, main statistics API have minimum aproximation of 1 minute, and values accuracy rises some time after, so default 5 minutes are reasonable good value for metrics scrape offset
```
### From env
For setup from env you need to set some specific variables, see example config above or
common.ClientConf struct for complete list). Configuration variables names comes from
yaml params full path in upper case with "SCE" prefix joined with "_" for eg. options above would be:
```
SCE_LISTEN_ADDRESS=":9921"
SCE_API_URL="https://api.cdn.sber.cloud"
SCE_API_USERNAME="username@example.com"
SCE_API_PASSWORD="password"
```

## Running in docker container

```
docker run --rm -e SCE_API_USERNAME="username@example.com" -e SCE_API_PASSWORD="password" -p 9921:9921/tcp ghcr.io/rabotaru/sbercdn-exporter:latest
```
