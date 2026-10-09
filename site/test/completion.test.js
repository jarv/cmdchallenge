import assert from 'node:assert/strict'
import { test } from 'node:test'
import { complete } from '../completion.js'

test('completes the word at the caret without changing the rest of a pipeline', () => {
  const result = complete('cat acc | wc -l', 7, ['access.log'])
  assert.equal(result.text, 'cat access.log | wc -l')
  assert.equal(result.caret, 14)
  assert.equal(complete('cat access | wc -l', 7, ['access.log']).text, 'cat access.log | wc -l')
  assert.equal(complete('cat "North P" | wc -l', 12, ['North Pole']).text, 'cat North\\ Pole | wc -l')
})

test('ambiguous completions insert only the shared prefix and deduplicate candidates', () => {
  const result = complete('gr', 2, ['grep', 'great', 'grep', 'cat'])
  assert.equal(result.text, 'gre')
  assert.deepEqual(result.matches, ['grep', 'great'])
})

test('quoted and escaped filenames are completed with shell-safe escaping', () => {
  for (const input of ["cat Santa\\'s", 'cat "Santa\'s']) {
    const result = complete(input, input.length, ["Santa's Workshop"])
    assert.equal(result.text, "cat Santa\\'s\\ Workshop")
  }
  const result = complete("cat 'North P", 12, ['North Pole'])
  assert.equal(result.text, 'cat North\\ Pole')
})

test('completion works after shell operators and leaves unmatched words alone', () => {
  assert.equal(complete('cat x|gr', 8, ['grep']).text, 'cat x|grep')
  assert.equal(complete('xyz', 3, ['grep']).text, 'xyz')
})
