// Identify the shell word at the caret, respecting quotes and escaped spaces.
// This is deliberately a word completer, not a shell parser or filesystem query.
export function completionToken (text, caret) {
  let start = 0
  let value = ''
  let quote = ''
  let escaped = false
  for (let index = 0; index < caret; index++) {
    const char = text[index]
    if (escaped) {
      value += char
      escaped = false
    } else if (char === '\\' && quote !== "'") {
      escaped = true
    } else if (quote) {
      if (char === quote) quote = ''
      else value += char
    } else if (char === "'" || char === '"') {
      quote = char
    } else if (/[\s|;&()<>]/.test(char)) {
      start = index + 1
      value = ''
    } else {
      value += char
    }
  }
  let end = caret
  for (; end < text.length; end++) {
    const char = text[end]
    if (escaped) escaped = false
    else if (char === '\\' && quote !== "'") escaped = true
    else if (quote) {
      if (char === quote) quote = ''
    } else if (char === "'" || char === '"') quote = char
    else if (/[\s|;&()<>]/.test(char)) break
  }
  return { start, end, value }
}

export function complete (text, caret, candidates) {
  const token = completionToken(text, caret)
  const matches = [...new Set(candidates)].filter(candidate => candidate.startsWith(token.value))
  if (!matches.length) return { text, caret, matches }
  let prefix = matches[0]
  for (const match of matches.slice(1)) {
    while (!match.startsWith(prefix)) prefix = prefix.slice(0, -1)
  }
  if (!prefix || prefix === token.value) return { text, caret, matches }
  // Backslash escaping works for names containing spaces, quotes, and shell operators.
  const replacement = prefix.replace(/([^\w./-])/g, '\\$1')
  return {
    text: text.slice(0, token.start) + replacement + text.slice(token.end),
    caret: token.start + replacement.length,
    matches
  }
}
