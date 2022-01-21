# sbercdn-exporter

Prometheus metrics exporter for SberCloud CDN https://cdn.sber.cloud.
API documentation https://docs.sbercloud.ru/cdn/ug/topics/guides__api.html

## Configuring
At this moment service may be configured from config and/or from env.
### From config file
There is only one cli flag now `-config` to get settings from YAML formatted config file for eg.:
```
listen:
  address: ":9921"

api:
  url: "https://api.cdn.sber.cloud"
  auth:
    username: "username@example.com"
    password: "password"
```
It is bare minimum, but you could configure more options, see common.ClientConf
### From env
For setup from env you need to set some specific variables (see example config sbercdn-exporter.yml or common.ClientConf for complete list). Configuration variables names comes from yaml params full path in upper case joined with "_" and prefix "SCE" joined with "_" for eg. options above would be:
```
SCE_LISTEN_ADDRESS=":9921"
SCE_API_URL="https://api.cdn.sber.cloud"
SCE_API_AUTH_USERNAME="username@example.com"
SCE_API_AUTH_PASSWORD="password"
```