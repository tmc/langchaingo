# 文档站点

本项目的文档使用 [Docusaurus](https://docusaurus.io/) 构建，它是一个现代化的静态网站生成器。

### 安装

```
$ npm i
```

### 本地开发

```
$ npm run start
```

此命令会启动一个本地开发服务器并在浏览器中打开一个窗口。大多数更改都会实时反映出来，无需重新启动服务器。

### 构建

```
$ npm run build
```

此命令会将静态内容生成到 `build` 目录中，并且可以使用任何静态内容托管服务进行部署。

### 部署

使用 SSH：

```
$ USE_SSH=true npm run deploy
```

不使用 SSH：

```
$ GIT_USER=<您的 GitHub 用户名> npm run deploy
```

如果您使用 GitHub Pages进行托管，此命令是一种构建网站并将更改推送到 `gh-pages` 分支的便捷方法。

### 持续集成

我们已经为您设置了一些常见的 linting/formatting (代码风格检查/格式化) 默认配置。如果您的项目与开源的持续集成系统（例如 Travis CI、CircleCI）集成，您可以使用以下命令检查问题。

```
$ npm run ci
```
