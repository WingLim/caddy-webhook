# caddy-webhook
Caddy v2 module for run webhook.

## Installation
```shell
xcaddy build \
  --with github.com/WingLim/caddy-webhook
```

## Usage

### Caddyfile

```
webhook [<url> <path>] {
    repo		<text>
    path 		<text>
    branch 		<text>
    depth		<int>
    type 		<text>
    secret		<text>
}
```

#### Example

```
route /webhook {
    webhook {
        repo github.com/WingLim/caddy-webhook
        path src/webhook
    }
}
```