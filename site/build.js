import { cp, mkdir, readFile, readdir, rm, writeFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { promisify } from 'node:util'
import { gzip } from 'node:zlib'
import esbuild from 'esbuild'
import * as sass from 'sass'
import chokidar from 'chokidar'

const root = path.dirname(fileURLToPath(import.meta.url))
const dist = path.join(root, 'dist')
const watching = process.argv.includes('--watch')
const compress = promisify(gzip)

const options = {
  absWorkingDir: root,
  entryPoints: ['main.js'],
  outdir: 'dist/assets',
  entryNames: '[name]-[hash]',
  bundle: true,
  format: 'esm',
  target: ['es2022'],
  minify: !watching,
  sourcemap: watching,
  metafile: true,
  write: false,
  external: ['/img/*', '/fonts/*'],
  plugins: [{
    name: 'sass',
    setup (build) {
      build.onLoad({ filter: /\.scss$/ }, async ({ path: filename }) => {
        const result = await sass.compileAsync(filename)
        return {
          contents: result.css,
          loader: 'css',
          resolveDir: path.dirname(filename),
          watchFiles: result.loadedUrls.filter(url => url.protocol === 'file:').map(fileURLToPath)
        }
      })
    }
  }]
}

async function gzipFiles (directory) {
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const filename = path.join(directory, entry.name)
    if (entry.isDirectory()) {
      await gzipFiles(filename)
    } else if (/\.(html|js|css|json|svg|txt)$/.test(entry.name)) {
      await writeFile(filename + '.gz', await compress(await readFile(filename)))
    }
  }
}

async function writeBuild (result) {
  await rm(dist, { recursive: true, force: true })
  await cp(path.join(root, 'public'), dist, { recursive: true })
  for (const file of result.outputFiles) {
    await mkdir(path.dirname(file.path), { recursive: true })
    await writeFile(file.path, file.contents)
  }
  const outputs = Object.entries(result.metafile.outputs)
  const [js, metadata] = outputs.find(([, info]) => info.entryPoint === 'main.js')
  const assetURL = filename => '/' + path.relative(dist, path.resolve(root, filename)).split(path.sep).join('/')
  const source = await readFile(path.join(root, 'index.html'), 'utf8')
  const html = source
    .replace('<!-- build:css -->', `<link rel="stylesheet" href="${assetURL(metadata.cssBundle)}">`)
    .replace('src="/main.js"', `src="${assetURL(js)}"`)
  await writeFile(path.join(dist, 'index.html'), html)
  if (!watching) await gzipFiles(dist)
  console.log(`Built site/dist${watching ? ' (development)' : ''}`)
}

async function main () {
  if (!watching) {
    await writeBuild(await esbuild.build(options))
    return
  }

  const context = await esbuild.context(options)
  const rebuild = async () => {
    try {
      await writeBuild(await context.rebuild())
    } catch (error) {
      console.error(error)
    }
  }
  await rebuild()
  let queue = Promise.resolve()
  let timer
  const watcher = chokidar.watch(root, {
    ignored: filename => /(^|[/\\])(node_modules|dist|test-results|playwright-report|\.git)([/\\]|$)/.test(filename),
    ignoreInitial: true
  }).on('all', () => {
    clearTimeout(timer)
    timer = setTimeout(() => { queue = queue.then(rebuild) }, 100)
  })
  watcher.on('ready', () => console.log('Watching site sources; refresh the browser after rebuilding.'))
  const shutdown = async () => {
    clearTimeout(timer)
    await watcher.close()
    await queue
    await context.dispose()
    process.exit(0)
  }
  process.once('SIGINT', shutdown)
  process.once('SIGTERM', shutdown)
}

main().catch(error => {
  console.error(error)
  process.exitCode = 1
})
