# Changelog

## Unreleased

- "Would I rather be feared or loved? Easy. Both. I want people to be afraid of how much they love me." — Michael Scott

## v1.14.0

- feat: Add support for Webpack 5 entry descriptors (#241)

## v1.13.0

- feat: Support minimal CLI options (#225)
- fix: Return an actual error for propagation (#224)
- deps: Bump sentry-cli to `1.58.0`

## v1.12.1

- fix(deploy): change deploy to newDeploy in mocked CLI object (#206)
- fix(types): add deploy configuration to type definitions (#208)

## v1.12.0

- feat: Allow to perform release deploys (#192)
- fix: CJS/TS Exports Interop (#190)
- fix: make setCommits.repo type optional (#200)
- deps: Bump sentry-cli to `1.55.0`

## v1.11.1

- meta: Bump sentry-cli to `1.52.3` which fixes output handlers

## v1.11.0

**This release sets `node.engine: >=8` which makes it incompatible with Node v6**
If you need to support Node v6, please pin your dependency to `1.10.0`
and use selective version resolution: https://classic.yarnpkg.com/en/docs/selective-version-resolutions/

- meta: Bump sentry-cli to `1.52.2`
- meta: Drop support for `node v6` due to new `sentry-cli` requiring `node >=8`
- chore: Fix setCommits types (#169)

## v1.10.0

- feat: Allow for skiping release finalization (#157)
- fix: Ensure afterEmit hook exists (#165)
- chore: Update TS definitions (#168)

## v1.9.3

- chore: Bump sentry-cli to `1.49.0`
- fix: Dont fail compilation if there is no release available (#155)
- fix: Update auto/repo logic for `setCommit` option (#156)

## v1.9.2

- chore: Resolve Snyk as dependency issues (#152)

## v1.9.1

- ref: Allow for nested setCommits (#142)
- fix: Fixed TS definitions export error (#145)

## v1.9.0

- feat: Add `setCommits` options (#139)
- chore: Add `TypeScript` definition file (#137)
- meta: Bump sentry-cli to `1.48.0`

## v1.8.1

- meta: Bump sentry-cli to `1.47.1`

## v1.8.0

- feat: Add errorHandler option (#133)

## v1.7.0

- feat: Add silent option to disable all output to stdout (#127)

## v1.6.2

- fix: Extract loader name in more reliable way
- build: Craft integration

## v1.6.1

- https://github.com/getsentry/sentry-webpack-plugin/releases
