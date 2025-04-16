# Docs site

The documentation for this project is built using [Docusaurus](https://docusaurus.io/), a modern static website generator.

### Installation

```
$ npm i
```

### Local Development

```
$ npm run start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

### Build

```
$ npm run build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

### Deployment

Using SSH:

```
$ USE_SSH=true npm run deploy
```

Not using SSH:

```
$ GIT_USER=<Your GitHub username> npm run deploy
```

If you are using GitHub pages for hosting, this command is a convenient way to build the website and push to the `gh-pages` branch.

### Continuous Integration

Some common defaults for linting/formatting have been set for you. If you integrate your project with an open source Continuous Integration system (e.g. Travis CI, CircleCI), you may check for issues using the following command.

```
$ npm run ci
```

