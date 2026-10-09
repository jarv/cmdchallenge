# CMD Challenge

This repository contains the code for the site [cmdchallenge.com](https://cmdchallenge.com)

[Read more about cmdchallenge](https://jarv.org/tags/cmdchallenge/)

## Installation

- [Install Docker](https://docs.docker.com/get-docker/)
- [Install rtx](https://github.com/jdxcode/rtx#quickstart)
- `rtx install`

```
docker-compose build runcmd cmd cmd-no-bin
# For ARM (M1 mac for example) run
#   BUILD_ARCH=arm64 docker-compose build runcmd cmd cmd-no-bin

docker-compose up --remove-orphans
# For ARM (M1 mac for example) run
#   BUILD_ARCH=arm64 docker-compose up runcmd caddy -V --remove-orphans

# Connect your browser to http://localhost:8181/
```

`cmd` and `cmd-no-bin` are runner images launched on demand by the Go server.
They use the `runner-images` Compose profile so a normal `docker-compose up`
starts only the web server and Caddy. Build them explicitly as shown above;
they are not long-running services.

- Connect your browser to http://localhost:8100

## Testing

- `cd cmdchallenge && go test ./...`
- Frontend checks (from `site/`):

  ```sh
  npm ci
  npm run lint
  npm test
  npm run build
  npx playwright install chromium
  npm run test:browser
  ```

  Browser tests exercise the production build on desktop and mobile with mocked
  command/solutions responses; they do not require Docker.

## Local development

### Static assets

```
cd site
npm ci
npm run build
```

`build.js` bundles browser JavaScript and CSS with esbuild, compiles Sass, copies
`public/`, and writes `dist/index.html` with content-hashed asset URLs. Production
builds include precompressed gzip files for Caddy.

For development, run `npm run dev` from `site/` to watch JavaScript, Sass, HTML,
and public assets. Start the Go server in another terminal and refresh the browser
after changes. Watch builds include source maps and skip minification/compression.
Restart the watcher after changing `build.js` or installing dependencies.

### Run the server

```
cd cmdchallenge
# Start the backend the `-dev` option uses an in-memory db.
go run cmd/runcmd/runcmd.go -dev -staticDistDir=../site/dist
```

Open <http://localhost:8181/>. The Go server serves both the built site and its API.
The other flavors can be tested at <http://oops.localhost:8181/> and
<http://12days.localhost:8181/> in browsers that resolve `.localhost` to loopback.

### Frontend structure

The frontend uses native DOM APIs and ES modules. `main.js` loads styles and starts
`app.js`, which coordinates challenge state, navigation, and requests. The command
form in `command-input.js` provides history, completion, keyboard handling, and
busy state. `ui.js` renders challenge content and feedback; `api.js`, `storage.js`,
and `router.js` handle the browser interfaces.

Enter runs a command, Shift+Enter inserts a new line, Up/Down recalls history, and
Tab completes a word. Double Tab displays matching candidates. Shift+Tab leaves
the command field, and Tab on an empty field moves to the Run button. Existing
`correct_answers` progress and jQuery Terminal history remain available.

Type `clear` (or press Ctrl+L) to clear the command field and displayed output
locally. Ctrl+A and Ctrl+E move to the start and end of the current line; Ctrl+C
cancels the current input when no text is selected, and copies selected text
normally. These shortcuts apply only while the command field has focus.

## Misc

**Test a single command:**

```
curl  http://localhost:8181/c/r -X POST -F slug=hello_world -F cmd="echo hello world"
```

**Fetch solutions:**

```
curl http://localhost:8181/c/s?slug=hello_world
```

## Bugs / Suggestions

- Open [a GitHub issue](https://github.com/jarv/cmdchallenge/-/issues).
