# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

### [15.0.2](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v15.0.1...v15.0.2) (2020-09-19)

## [15.0.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v14.1.1...v15.0.0) (2020-09-19)

Support new Edge browser, updated babel plugins. Modern browsers are now: Edge >= 83, Firefox >= 78, Chrome >= 80, Opera >= 67, Safari >= 13.1, iOS >= 13.4


## [14.1.1](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v14.1.0...v14.1.1) (2019-12-14)


### Bug Fixes

* require default in both added plugins ([b75a8a7](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/b75a8a7))



# [14.1.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v14.0.0...v14.1.0) (2019-12-13)


### Features

* add optional chaining and nullish coalescing operator ([df79b4b](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/df79b4b))



# [14.0.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v13.1.0...v14.0.0) (2019-04-05)


### Features

* es2019 and json strings ([78345b2](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/78345b2))


### BREAKING CHANGES

* when edge: false, requires firefox 62 and safari 12



# [13.1.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v13.0.1...v13.1.0) (2019-03-09)


### Bug Fixes

* asynchronous Iterators safari 12 ([e5f4ff8](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/e5f4ff8))


### Features

* update dependencies ([9345d02](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/9345d02))



## [13.0.1](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v13.0.0...v13.0.1) (2018-11-24)



# [13.0.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v12.0.0...v13.0.0) (2018-11-24)


### Features

* add supportVariablesFunctionName ([5064f23](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/5064f23))
* pass optional catch binding in modern browsers ([c9b6af5](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/c9b6af5))


### BREAKING CHANGES

* when edge:false, modern browsers are now: firefox 58, chrome 66, opera 53, safari 11.1



# [12.0.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v12.0.0-beta.1...v12.0.0) (2018-08-28)



# [12.0.0-beta.1](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v12.0.0-beta.0...v12.0.0-beta.1) (2018-04-27)


### chore

* update dependencies ([94b08a5](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/94b08a5))


### BREAKING CHANGES

* drop node 4



# [12.0.0-beta.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v11.0.1...v12.0.0-beta.0) (2018-04-06)


### Features

* babel 7 and shipped proposals ([1dda800](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/1dda800))



## [11.0.1](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v11.0.0...v11.0.1) (2018-04-06)



# [11.0.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v10.0.1...v11.0.0) (2018-04-06)


### Code Refactoring

* remove buildPreset compatibility ([9a7b01e](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/9a7b01e))


### Features

* option es2018, remove option esnext and safari 10 ([e5836f5](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/e5836f5))


### BREAKING CHANGES

* removed buildPreset
* option esnext and safari 10 removed



## [10.0.1](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v10.0.0...v10.0.1) (2017-10-22)



# [10.0.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v9.0.2...v10.0.0) (2017-08-15)


### Features

* add dynamic import syntax ([d0e7a7e](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/d0e7a7e))
* add object rest/spread ([d3c695f](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/d3c695f))
* enable `edge` and `safari10` by default ([c3212e7](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/c3212e7))
* firefox 53, safari 10.1 and edge 15 ([b66b6cd](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/b66b6cd))



## [9.0.2](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v9.0.1...v9.0.2) (2017-03-09)



## [9.0.1](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v9.0.0...v9.0.1) (2017-03-09)


### Bug Fixes

* reenable transform-async-to-generator ([7b26e8f](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/7b26e8f))



# [9.0.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v8.1.2...v9.0.0) (2017-03-08)


### Features

* async functions and exponentiation operator ([1f1682e](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/1f1682e))


### BREAKING CHANGES

* drop support firefox < 42, chrome < 55, opera < 42, safari < 10.1, option `safari10`



## [8.1.2](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v8.1.1...v8.1.2) (2017-03-03)



## [8.1.1](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v8.1.0...v8.1.1) (2017-02-27)


### Bug Fixes

* missing dependencies ([f925a55](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/f925a55)), closes [#16](https://github.com/christophehurpeau/babel-preset-modern-browsers/issues/16)



# [8.1.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v8.0.0...v8.1.0) (2017-02-25)


### Features

* add es2017 plugins and es2016, es2017 option ([ade31a8](https://github.com/christophehurpeau/babel-preset-modern-browsers/commit/ade31a8))



# [8.0.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v7.0.0...v8.0.0) (2017-01-24)



# [7.0.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v6.0.0...v7.0.0) (2016-11-17)



# [6.0.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v5.1.0...v6.0.0) (2016-10-10)



# [5.1.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v5.0.2...v5.1.0) (2016-08-01)



## [5.0.2](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v5.0.0...v5.0.2) (2016-06-27)



# [5.0.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v4.1.0...v5.0.0) (2016-06-27)



# [4.1.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v3.0.0...v4.1.0) (2016-06-23)



# [3.0.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v2.1.1...v3.0.0) (2016-06-11)



## [2.1.1](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v2.1.0...v2.1.1) (2016-06-03)



# [2.1.0](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v2.0.1...v2.1.0) (2016-05-26)



## [2.0.1](https://github.com/christophehurpeau/babel-preset-modern-browsers/compare/v2.0.0...v2.0.1) (2016-05-20)



# 2.0.0 (2016-04-26)
