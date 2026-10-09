import assert from 'node:assert/strict'
import { test } from 'node:test'
import { runCommand, fetchSolutions } from '../api.js'

test('command requests preserve UTF-8 and the backend form contract', async t => {
  t.mock.method(globalThis, 'fetch', async (url, options) => {
    assert.equal(url, '/c/r')
    assert.equal(options.method, 'POST')
    assert.equal(Buffer.from(options.body.get('cmd'), 'base64').toString('utf8'), 'echo café 🐚')
    assert.equal(options.body.get('slug'), 'hello_world')
    assert.equal(options.body.get('version'), '1')
    return new Response(JSON.stringify({ Correct: true, ExitCode: 0 }))
  })
  await runCommand('echo café 🐚', { slug: 'hello_world', version: 1 })
})

test('HTTP errors preserve the backend error message', async t => {
  t.mock.method(globalThis, 'fetch', async () => new Response('Slow down!', { status: 429 }))
  await assert.rejects(runCommand('ls', { slug: 'hello_world', version: 1 }), /Slow down!/)
})

test('malformed runner and solutions responses are rejected', async t => {
  t.mock.method(globalThis, 'fetch', async () => new Response('{}'))
  await assert.rejects(runCommand('ls', { slug: 'hello_world', version: 1 }), /Invalid response/)
  await assert.rejects(fetchSolutions('hello_world'), /Invalid solutions response/)
})
