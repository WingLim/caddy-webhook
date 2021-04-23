# caddy-webhook
Caddy v2 module for server a webhook.

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
    command <text>...
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
- **command** - the command run when repo initializes or get the correct webhook request

#### Example

The full example to run a hugo blog:
```
example.com

root * blog/www
file_server

route /webhook {
    webhook {
        repo https://github.com/WingLim/winglim.github.io.git
        path blog
        branch hugo
        command hugo --destination www
        submodule   
    }
}
```
This `Caddyfile` will clone https://github.com/WingLim/winglim.github.io.git to directory `blog` when initializes.

Then run the command `hugo --destination www` inside the `blog` to generate our blog to `www`.

When receive webhook request in `/webhook` which we set it route, will update the repo and run command again.