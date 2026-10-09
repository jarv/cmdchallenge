import assert from 'node:assert/strict'
import { test } from 'node:test'
import { spawn } from 'node:child_process'
import { cp, mkdtemp, readFile, readdir, rm, symlink, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { gunzipSync } from 'node:zlib'

const root = fileURLToPath(new URL('..', import.meta.url))

async function fixture () {
  const directory = await mkdtemp(path.join(tmpdir(), 'cmdchallenge-build-'))
  await cp(root, directory, {
    recursive: true,
    filter: filename => !['node_modules', 'dist', 'test-results', 'playwright-report', '.git'].includes(path.basename(filename))
  })
  await symlink(path.join(root, 'node_modules'), path.join(directory, 'node_modules'), 'dir')
  return directory
}

function runBuild (directory, args = []) {
  const child = spawn(process.execPath, ['build.js', ...args], { cwd: directory })
  let output = ''
  child.stdout.on('data', data => { output += data })
  child.stderr.on('data', data => { output += data })
  return { child, output: () => output }
}

async function waitFor (predicate, description) {
  const deadline = Date.now() + 15000
  while (!await predicate()) {
    if (Date.now() > deadline) throw new Error('Timed out: ' + description)
    await new Promise(resolve => setTimeout(resolve, 50))
  }
}

test('production build emits working asset references and valid gzip sidecars', async t => {
  const directory = await fixture()
  t.after(() => rm(directory, { recursive: true, force: true }))
  const build = runBuild(directory)
  const status = await new Promise(resolve => build.child.on('close', resolve))
  assert.equal(status, 0, build.output())
  const html = await readFile(path.join(directory, 'dist/index.html'), 'utf8')
  assert.doesNotMatch(html, /jquery|\/main\.js|build:css/)
  const assets = [...html.matchAll(/(?:src|href)="(\/assets\/[^" ]+)"/g)].map(match => match[1])
  assert.equal(assets.length, 2)
  for (const asset of ['/index.html', ...assets]) {
    const content = await readFile(path.join(directory, 'dist' + asset))
    const compressed = await readFile(path.join(directory, 'dist' + asset + '.gz'))
    assert.deepEqual(gunzipSync(compressed), content)
  }
  assert.deepEqual(await readFile(path.join(directory, 'dist/img/cmdchallenge-round.png')), await readFile(path.join(root, 'public/img/cmdchallenge-round.png')))
})

test('watch rebuilds imported Sass, HTML, and added/removed public assets', async t => {
  const directory = await fixture()
  const stylesheet = path.join(directory, 'sass/command-input.scss')
  const partial = path.join(directory, 'sass/_watch.scss')
  await writeFile(stylesheet, '@use "watch";\n' + await readFile(stylesheet, 'utf8'))
  await writeFile(partial, '.watch-check { color: red; }')
  const build = runBuild(directory, ['--watch'])
  t.after(async () => {
    if (build.child.exitCode === null && build.child.signalCode === null) {
      const closed = new Promise(resolve => build.child.on('close', resolve))
      build.child.kill('SIGTERM')
      await closed
    }
    await rm(directory, { recursive: true, force: true })
  })
  await waitFor(() => build.output().includes('Watching site sources'), build.output())
  await writeFile(partial, '.watch-check { color: blue; }')
  await waitFor(async () => {
    try {
      const assets = await readdir(path.join(directory, 'dist/assets'))
      const css = await readFile(path.join(directory, 'dist/assets', assets.find(asset => asset.endsWith('.css'))), 'utf8')
      return css.includes('color: blue')
    } catch { return false }
  }, 'Sass dependency rebuild')
  const source = path.join(directory, 'index.html')
  await writeFile(source, (await readFile(source, 'utf8')).replace('<body>', '<body data-watch="updated">'))
  await waitFor(async () => {
    try { return (await readFile(path.join(directory, 'dist/index.html'), 'utf8')).includes('data-watch="updated"') } catch { return false }
  }, 'HTML rebuild')
  const asset = path.join(directory, 'public/watch-check.txt')
  await writeFile(asset, 'updated asset')
  await waitFor(async () => {
    try { return (await readFile(path.join(directory, 'dist/watch-check.txt'), 'utf8')) === 'updated asset' } catch { return false }
  }, 'public asset added')
  await rm(asset)
  await waitFor(async () => {
    try { return !(await readdir(path.join(directory, 'dist'))).includes('watch-check.txt') } catch { return false }
  }, 'public asset removed')
})
