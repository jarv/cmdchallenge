# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

### [1.0.3](https://github.com/webpack-contrib/expose-loader/compare/v1.0.2...v1.0.3) (2020-11-26)


### Bug Fixes

* set side effects to false ([#122](https://github.com/webpack-contrib/expose-loader/issues/122)) ([ee2631d](https://github.com/webpack-contrib/expose-loader/commit/ee2631df243e4fa13f107189be5dc469108495b3))

### [1.0.2](https://github.com/webpack-contrib/expose-loader/compare/v1.0.1...v1.0.2) (2020-11-25)


### Bug Fixes

* don't strip loader "ref" from import string ([6271fc4](https://github.com/webpack-contrib/expose-loader/commit/6271fc4e227a63aae082b9a111e103b6967bc1ba))

### [1.0.1](https://github.com/webpack-contrib/expose-loader/compare/v1.0.0...v1.0.1) (2020-10-09)

### Chore

* update `schema-utils`

## [1.0.0](https://github.com/webpack-contrib/expose-loader/compare/v0.7.5...v1.0.0) (2020-06-23)


### ⚠ BREAKING CHANGES

* minimum supported Node.js version is `10.13`
* minimum supported `webpack` version is `4`
* `inline` syntax was changed, please [read](https://github.com/webpack-contrib/expose-loader#inline)
* list of exposed values moved to the `exposes` option, please [read](https://github.com/webpack-contrib/expose-loader#exposes)
* migrate away from `pitch` phase
* do not override existing exposed values in the global object by default, because it is unsafe, please [read](https://github.com/webpack-contrib/expose-loader#override)

### Features

* validate options
* support webpack 5
* support multiple exposed values
* interpolate exposed values
* allow to expose part of a module
* allow to expose values with `.` (dot) in the name

### Fixes

* do not break source maps
* do not generate different hashed on different os
* compatibility with ES module syntax

<a name="0.7.5"></a>
## [0.7.5](https://github.com/webpack-contrib/expose-loader/compare/v0.7.4...v0.7.5) (2018-03-09)


### Bug Fixes

* **package:** add `webpack >= v4.0.0` (`peerDependencies`) ([#67](https://github.com/webpack-contrib/expose-loader/issues/67)) ([daf39ea](https://github.com/webpack-contrib/expose-loader/commit/daf39ea))



<a name="0.7.4"></a>
## 0.7.4 (2017-11-18)


### Bug Fixes

* **hash:** inconsistent hashes for builds in different dirs. ([#28](https://github.com/webpack-contrib/expose-loader/issues/28)) ([efe59de](https://github.com/webpack-contrib/expose-loader/commit/efe59de))
* **remainingRequest:** resolve  issue when multiple variables are exposed for the same request. ([#30](https://github.com/webpack-contrib/expose-loader/issues/30)) ([335f9e6](https://github.com/webpack-contrib/expose-loader/commit/335f9e6))
* ensure `userRequest` stays unique (`module.userRequest`) ([#58](https://github.com/webpack-contrib/expose-loader/issues/58)) ([51629a4](https://github.com/webpack-contrib/expose-loader/commit/51629a4))
