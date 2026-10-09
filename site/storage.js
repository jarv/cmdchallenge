const PROGRESS_KEY = 'correct_answers'
const HISTORY_KEY = 'cmdchallenge_history'

export function readArray (key) {
  try {
    const value = JSON.parse(localStorage.getItem(key))
    return Array.isArray(value) ? value.filter(item => typeof item === 'string') : []
  } catch {
    return []
  }
}

function writeArray (key, value) {
  try {
    localStorage.setItem(key, JSON.stringify(value))
  } catch {
    // Keep the current session usable when browser storage is unavailable.
  }
}

export function loadProgress () {
  return new Set(readArray(PROGRESS_KEY))
}

export function saveProgress (progress) {
  writeArray(PROGRESS_KEY, [...progress])
}

export function loadHistory () {
  return readArray(HISTORY_KEY).length
    ? readArray(HISTORY_KEY)
    : readArray('cmdchallenge_0_commands')
}

export function saveHistory (history) {
  writeArray(HISTORY_KEY, history.slice(-500))
}
