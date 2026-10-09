import assert from 'node:assert/strict'
import { test } from 'node:test'
import { loadProgress, saveProgress, loadHistory } from '../storage.js'

function mockStorage (t, initial) {
  const values = new Map(Object.entries(initial))
  globalThis.localStorage = {
    getItem: key => values.get(key) ?? null,
    setItem: (key, value) => values.set(key, value)
  }
  t.after(() => { delete globalThis.localStorage })
  return values
}

test('existing progress is preserved when another challenge is completed', t => {
  const values = mockStorage(t, { correct_answers: '["hello_world"]' })
  const progress = loadProgress()
  progress.add('current_working_directory')
  saveProgress(progress)
  assert.deepEqual(JSON.parse(values.get('correct_answers')), ['hello_world', 'current_working_directory'])
})

test('corrupt and unavailable storage do not prevent a session from working', t => {
  mockStorage(t, { correct_answers: '{}' })
  assert.equal(loadProgress().size, 0)
  localStorage.getItem = () => { throw new Error('Storage unavailable') }
  localStorage.setItem = () => { throw new Error('Storage unavailable') }
  assert.equal(loadProgress().size, 0)
  assert.doesNotThrow(() => saveProgress(new Set(['hello_world'])))
})

test('existing jQuery Terminal history is available after migration', t => {
  mockStorage(t, { cmdchallenge_0_commands: '["echo hello world","pwd"]' })
  assert.deepEqual(loadHistory(), ['echo hello world', 'pwd'])
})
