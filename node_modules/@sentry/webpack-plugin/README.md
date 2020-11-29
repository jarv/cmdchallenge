<p align="center">
    <a href="https://sentry.io" target="_blank" align="center">
        <img src="https://sentry-brand.storage.googleapis.com/sentry-logo-black.png" width="280">
    </a>
<br/>
    <h1>Sentry Webpack Plugin</h1>
</p>

[![Travis](https://img.shields.io/travis/getsentry/sentry-webpack-plugin.svg?maxAge=2592000)](https://travis-ci.org/getsentry/sentry-webpack-plugin)
[![codecov](https://codecov.io/gh/getsentry/sentry-webpack-plugin/branch/master/graph/badge.svg)](https://codecov.io/gh/getsentry/sentry-webpack-plugin)
[![npm version](https://img.shields.io/npm/v/@sentry/webpack-plugin.svg)](https://www.npmjs.com/package/@sentry/webpack-plugin)
[![npm dm](https://img.shields.io/npm/dm/@sentry/webpack-plugin.svg)](https://www.npmjs.com/package/@sentry/webpack-plugin)
[![npm dt](https://img.shields.io/npm/dt/@sentry/webpack-plugin.svg)](https://www.npmjs.com/package/@sentry/webpack-plugin)

[![deps](https://david-dm.org/getsentry/sentry-webpack-plugin/status.svg)](https://david-dm.org/getsentry/sentry-webpack-plugin?view=list)
[![deps dev](https://david-dm.org/getsentry/sentry-webpack-plugin/dev-status.svg)](https://david-dm.org/getsentry/sentry-webpack-plugin?type=dev&view=list)
[![deps peer](https://david-dm.org/getsentry/sentry-webpack-plugin/peer-status.svg)](https://david-dm.org/getsentry/sentry-webpack-plugin?type=peer&view=list)

A webpack plugin acting as an interface to
[Sentry CLI](https://docs.sentry.io/learn/cli/).

### Installation

Using npm:

```
$ npm install @sentry/webpack-plugin --save-dev
```

Using yarn:

```
$ yarn add @sentry/webpack-plugin --dev
```

### CLI Configuration

You can use either `.sentryclirc` file or ENV variables described here
https://docs.sentry.io/cli/configuration.

### Usage

```js
const SentryCliPlugin = require('@sentry/webpack-plugin');

const config = {
  plugins: [
    new SentryCliPlugin({
      include: '.',
      ignoreFile: '.sentrycliignore',
      ignore: ['node_modules', 'webpack.config.js'],
      configFile: 'sentry.properties',
    }),
  ],
};
```

Also, check the [example](example) directory.

#### Options

| Option | Type | Required | Description |
---------|------|----------|-------------
| include | `string`/`array` | required | One or more paths that Sentry CLI should scan recursively for sources. It will upload all `.map` files and match associated `.js` files. |
| org | `string` | optional | The slug of the Sentry organization associated with the app. |
| project | `string` | optional | The slug of the Sentry project associated with the app. |
| authToken | `string` | optional | The authentication token to use for all communication with Sentry. Can be obtained from https://sentry.io/settings/account/api/auth-tokens/. |
| url | `string` | optional | The base URL of your Sentry instance. Defaults to https://sentry.io/, which is the correct value for SAAS customers. |
| vcsRemote | `string` | optional | The name of the remote in the version control system. Defaults to `origin`. |
| release | `string` | optional | Unique identifier for the release. Defaults to the output of the `sentry-cli releases propose-version` command, which automatically detects values for Cordova, Heroku, AWS CodeBuild, CircleCI, Xcode, and Gradle, and otherwise uses `HEAD`'s commit SHA. (**For `HEAD` option, requires access to `git` CLI and for the root directory to be a valid repository**). |
| dist | `string` | optional | Unique identifier for the distribution, used to further segment your release. Usually your build number. |
| entries | `array`/`RegExp`/`function(key: string): bool` | optional | Filter for entry points that should be processed. By default, the release will be injected into all entry points. |
| ignoreFile | `string` | optional | Path to a file containing list of files/directories to ignore. Can point to `.gitignore` or anything with the same format. |
| ignore | `string`/`array` | optional | One or more paths to ignore during upload. Overrides entries in `ignoreFile` file. If neither `ignoreFile` nor `ignore` is present, defaults to `['node_modules']`. |
| configFile | `string` | optional | Path to Sentry CLI config properties, as described in https://docs.sentry.io/product/cli/configuration/#configuration-file. By default, the config file is looked for upwards from the current path, and defaults from `~/.sentryclirc` are always loaded |
| ext | `array` | optional | The file extensions to be considered. By default the following file extensions are processed: `js`, `map`, `jsbundle`, and `bundle`. |
| urlPrefix | `string` | optional | URL prefix to add to the beginning of all filenames. Defaults to `~/` but you might want to set this to the full URL. This is also useful if your files are stored in a sub folder. eg: `url-prefix '~/static/js'`. |
| urlSuffix | `string` | optional | URL suffix to add to the end of all filenamess. Useful for appending query parameters. |
| validate | `boolean` | optional | When `true`, attempts source map validation before upload if rewriting is not enabled. It will spot a variety of issues with source maps and cancel the upload if any are found. Defaults to `false` to prevent false positives canceling upload. |
| stripPrefix | `array` | optional | When paired with `rewrite`, will remove a prefix from uploaded filenames. Useful for removing a path that is build-machine-specific. |
| stripCommonPrefix | `boolean` | optional |  When paired with `rewrite`, will add `~` to the `stripPrefix` array. Defaults to `false`.|
| sourceMapReference | `boolean` | optional | Prevents the automatic detection of sourcemap references. Defaults to `false`.|
| rewrite | `boolean` | optional | Enables rewriting of matching source maps so that indexed maps are flattened and missing sources are inlined if possible. Defaults to `true` |
| finalize | `boolean` | optional | Determines whether Sentry release record should be automatically finalized (`date_released` timestamp added) after artifact upload. Defaults to `true` |
| dryRun | `boolean` | optional | Attempts a dry run (useful for dev environments). Defaults to `false`. |
| debug | `boolean` | optional | Print useful debug information. Defaults to `false`.|
| silent | `boolean` | optional | Suppresses all logs (useful for `--json` option). Defaults to `false`. |
| errorHandler | `function(err: Error, invokeErr: function(): void, compilation: Compilation): void` | optional | Function to call a when CLI error occurs. Webpack compilation failure can be triggered by calling `invokeErr` callback. Can emit a warning rather than an error (allowing compilation to continue) by setting this to `(err, invokeErr, compilation) => { compilation.warnings.push('Sentry CLI Plugin: ' + err.message) }`. Defaults to `(err, invokeErr) => { invokeErr() }`. |
| setCommits | `Object` | optional | Adds commits to Sentry. See [table below](#setCommits) for details. |
| deploy | `Object` | optional | Creates a new release deployment in Sentry. See [table below](#deploy) for details. |


#### <a name="setCommits"></a>options.setCommits:

| Option | Type | Required | Description |
---------|------|----------|-------------
| repo | `string` | see notes | The full git repo name as defined in Sentry. Required if `auto` option is not `true`, otherwise optional. |
| commit | `string` | see notes | The current (most recent) commit in the release. Required if `auto` option is not `true`, otherwise optional. |
| previousCommit | `string` | optional | The last commit of the previous release. Defaults to the most recent commit of the previous release in Sentry, or if no previous release is found, 10 commits back from `commit`. |
| auto | `boolean` | optional | Automatically set `commit` and `previousCommit`. Defaults `commit` to `HEAD` and `previousCommit` as described above. Overrides other options |

#### <a name="deploy"></a>options.deploy:

| Option | Type | Required | Description |
---------|------|----------|-------------
| env | `string` | required | Environment value for the release, for example `production` or `staging`. |
| started | `number` | optional | UNIX timestamp for deployment start. |
| finished | `number` | optional | UNIX timestamp for deployment finish. |
| time | `number` | optional | Deployment duration in seconds. Can be used instead of `started` and `finished`. |
| name | `string` | optional | Human-readable name for this deployment. |
| url | `string` | optional | URL that points to the deployment. |

You can find more information about these options in our official docs:
https://docs.sentry.io/product/cli/releases/#sentry-cli-sourcemaps.
