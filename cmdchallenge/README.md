# CMDChallenge

This will be the rewrite of cmdchallenge to use a golang service in place of the serviceless components which currently consist of:

- AWS Lambda
- API Gateway
- Cloudflare
- S3 object storage

This re-write will move all cloud components to a single VM

## Tasks for rewrite

- [x] Accept POST with
  - challenge_slug
  - cmd
- [x] Check for command length < 300
- [x] Check for `X-Forwarded-For` header (not applicable for now)
- [x] Split command using `shlex.split`
- [x] See if the challenge is valid
- [x] Create fingerprint
- [x] Execute command using Docker locally
- [x] Test locally with browser
- [x] Add test suite
- [x] Allow a specific image tag to be specified
- [x] Store result in DB
- [x] Add caching using DB lookup
- [x] Add tests for caching
- [x] Return a single Error instead of Errors
- [x] Add SQLite data store
- [x] Replace python json generation
- [x] Remove Python code
- [x] Add application metrics
- [x] Deployment and config
  - [x] Adapt CI Pipeline
  - [x] Configure service on instance
  - [x] Deploy binary and prometheus
  - [x] Make userdate more idempotent
  - [x] Cron for SQLite DB backup
  - [x] Cut AMI
- [x] Validate on testing
- [x] Implement rate limiting https://github.com/didip/tollbooth
- [x] Validate test failure response on testing
- [x] Validate random failure response on testing
- [x] Add cmd to db
- [x] Seed the testing db with solutions
- [x] Disable networking / metadata
- [x] Pull images
- [x] Create /s for solutions
- [x] Metrics for correct/incorrect (answer=correct|incorrect)
- [x] Lower timeout for container
- [x] Create cleanup service
- [x] Update README.md
- [x] Update dashboard
- [x] Deploy to Production
- [x] Create testing/prod images instead of sha

## References

- https://github.com/eliben/code-for-blog/blob/master/2021/go-rest-servers/stdlib-basic/internal/taskstore/taskstore.go
- https://github.com/ardanlabs/python-go/blob/master/sqlite/trades/trades.go
