# caddy-webhook
Caddy v2 module for listening webhook.

## Installation
```shell
xcaddy build \
  --with github.com/WingLim/caddy-webhook
```

## Usage

### Caddyfile

```
webhook [<url> <path>] {
    repo	<text>
    path 	<text>
    branch 	<text>
    depth	<int>
    type 	<text>
    secret	<text>
    submodule
}
```

- **repo** - git repository url.
- **path** - path to clone and update repository.
- **branch** - branch to pull. Default is `main`.
- **depth** - depth for pull. Default is `0`.
- **type** - webhook type. Default is `github`.
- **secret** - secret to verify webhook request.
- **submodule** - enable recurse submodules

#### Example

```
route /webhook {
    webhook {
        repo github.com/WingLim/caddy-webhook
        path src/webhook
    }
}
```