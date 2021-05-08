# caddy-webhook

[![Build](https://github.com/WingLim/caddy-webhook/actions/workflows/build.yml/badge.svg)](https://github.com/WingLim/caddy-webhook/actions/workflows/build.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

用于服务 webhook 请求的 Caddy v2 模块。 

[English](https://github.com/WingLim/caddy-webhook/blob/main/README.md) | 中文

## 安装

### 使用 xcaddy 构建

```shell
xcaddy build \
  --with github.com/WingLim/caddy-webhook
```

### 在 Docker 中运行

在 [caddy-docker](https://github.com/WingLim/caddy-docker) 中查看 `Dockerfile`.

DockerHub: [winglim/caddy](https://hub.docker.com/repository/docker/winglim/caddy)

GitHub Package: [winglim/caddy](http://ghcr.io/winglim/caddy)

## 使用方法

支持的 WebHook 类型:

- github
- gitlab
- gitee
- bitbucket
- gogs

### Caddyfile 格式

注意：`webhook` 要作为 `rotue` 的最后一个 handler，因为 `caddy-webhook` 处理完请求后返回 `nil` 而不是执行下一个中间件。
所以放在 `webhook` 后的 handler 都不会生效。

```
webhook [<repo> <path>] {
    repo       <text>
    path       <text>
    branch     <text>
    depth      <int>
    type       <text>
    secret     <text>
    command    <text>...
    key	       <text>
    username   <text>
    password   <text>
    token      <text>
    submodule
}
```

- **repo** - git 仓库地址，支持 http、https和ssh。
- **path** - git 仓库的本地路径。
- **branch** - 分支名。默认值为 `main`。
- **depth** - pull 操作时的深度。 默认值为 `0`。
- **type** - webhook 类型. 默认值为 `github`.
- **secret** - 用于验证 webhook 请求。
- **submodule** - 是否拉取子模块。
- **command** - 初始化以及收到合法的 webhook 请求后执行的命令。
- **key** - 通过 ssh 获取 git 仓库时所需的私钥地址。
- **username** - 用于 http 验证的用户名。
- **password** - 用于 http 验证的密码。
- **token** - GitHub 个人授权 token。

### 样例

一个运行 hugo 博客的完整样例:

`Caddyfile`:

```
example.com

root www
file_server

route /webhook {
    webhook {
        repo https://github.com/WingLim/winglim.github.io.git
        path blog
        branch hugo
        command hugo --destination ../www
        submodule   
    }
}
```

在上面的配置中，webhook 模块会执行如下步骤：

1. 初始化时克隆 https://github.com/WingLim/winglim.github.io.git 到 `blog`。

    1. 如果仓库以及存在，那么会更新并切换到你设置的分支。

2. 在 `blog` 目录下执行 `hugo --destination ../www`。

3. 在 `/webhook` 监听并处理 webhook 请求。
    1. 接收到合法的 webhook 请求后，会再次执行第2步。

## 感谢

- [caddygit](https://github.com/vrongmeal/caddygit) - Git module for Caddy v2
- [caddy-git](https://github.com/abiosoft/caddy-git) - git middleware for Caddy
- [caddy-exec](https://github.com/abiosoft/caddy-exec) - Caddy v2 module for running one-off commands