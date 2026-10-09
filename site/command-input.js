import { complete } from './completion.js'
import { loadHistory, saveHistory } from './storage.js'
import { BlockCursor } from './block-cursor.js'

export class CommandInput {
  constructor ({ onSubmit, onClear, candidates }) {
    this.form = document.getElementById('command-form')
    this.input = document.getElementById('command-input')
    this.cursor = new BlockCursor(this.input)
    this.prompt = document.getElementById('command-exit-code')
    this.button = this.form.querySelector('button')
    this.completions = document.getElementById('completions')
    this.candidates = () => ['clear', ...candidates()]
    this.history = loadHistory()
    this.historyIndex = this.history.length
    this.draft = ''
    this.busy = false
    this.lastTab = 0
    this.form.addEventListener('submit', event => {
      event.preventDefault()
      if (this.busy) return
      const command = this.input.value.trim()
      this.input.value = ''
      this.resize()
      this.hideCompletions()
      this.historyIndex = this.history.length
      this.draft = ''
      if (!command || command === 'clear') {
        this.setExitCode(0)
        onClear()
        this.focus()
        return
      }
      if (this.history.at(-1) !== command) this.history.push(command)
      this.history = this.history.slice(-500)
      saveHistory(this.history)
      this.historyIndex = this.history.length
      this.draft = ''
      onSubmit(command)
    })
    this.input.addEventListener('input', () => {
      this.lastTab = 0
      this.hideCompletions()
      this.resize()
    })
    this.input.addEventListener('keydown', event => {
      if (event.isComposing || this.busy) return
      if (event.key !== 'Tab') this.lastTab = 0
      if (event.key === 'Enter' && !event.shiftKey) {
        event.preventDefault()
        this.form.requestSubmit()
      } else if (event.key === 'ArrowUp' || event.key === 'ArrowDown') {
        // Keep native vertical movement available inside pasted multiline commands.
        if (this.input.value.includes('\n')) return
        event.preventDefault()
        if (this.historyIndex === this.history.length) this.draft = this.input.value
        const step = event.key === 'ArrowUp' ? -1 : 1
        this.historyIndex = Math.max(0, Math.min(this.history.length, this.historyIndex + step))
        this.input.value = this.historyIndex === this.history.length ? this.draft : this.history[this.historyIndex]
        this.input.setSelectionRange(this.input.value.length, this.input.value.length)
        this.resize()
      } else if (event.key === 'Tab' && !event.shiftKey) {
        const now = Date.now()
        // Empty Tab and Shift+Tab retain normal keyboard focus navigation.
        if (!this.input.value) return
        event.preventDefault()
        const result = complete(this.input.value, this.input.selectionStart, this.candidates())
        this.input.value = result.text
        this.input.setSelectionRange(result.caret, result.caret)
        if (now - this.lastTab < 1000) this.showCompletions(result.matches)
        this.lastTab = now
        this.resize()
      } else if (event.key === 'Escape') {
        this.hideCompletions()
      } else if (event.ctrlKey && !event.altKey && !event.metaKey && (event.key === 'a' || event.key === 'e')) {
        event.preventDefault()
        const value = this.input.value
        const caret = this.input.selectionStart
        const start = caret === 0 ? 0 : value.lastIndexOf('\n', caret - 1) + 1
        const newline = value.indexOf('\n', caret)
        const end = newline === -1 ? value.length : newline
        const position = event.key === 'a' ? start : end
        this.input.setSelectionRange(position, position)
        this.cursor.update({ reveal: true })
      } else if (event.ctrlKey && (event.key === 'c' || event.key === 'l')) {
        // Preserve Ctrl+C copying when text is selected.
        if (event.key === 'c' && this.input.selectionStart !== this.input.selectionEnd) return
        event.preventDefault()
        this.input.value = ''
        this.resize()
        this.hideCompletions()
        if (event.key === 'l') {
          this.setExitCode(0)
          onClear()
        }
      }
    })
    this.setExitCode(0)
    window.addEventListener('resize', () => this.resize())
  }

  resize () {
    this.input.style.height = 'auto'
    this.input.style.height = `${this.input.scrollHeight}px`
    this.cursor.update()
  }

  focus () {
    this.input.focus({ preventScroll: true })
  }

  setExitCode (code) {
    this.prompt.textContent = code === null ? '☠️' : String(code)
    this.prompt.classList.toggle('failed', code !== 0)
  }

  setBusy (busy) {
    this.busy = busy
    this.input.readOnly = busy
    this.button.disabled = busy
    this.form.hidden = busy
    document.getElementById('term-challenge').classList.toggle('running', busy)
    this.form.setAttribute('aria-busy', String(busy))
    document.getElementById('term-spinner').hidden = !busy
    this.cursor.update()
    if (!busy) this.focus()
  }

  showCompletions (matches) {
    clearTimeout(this.completionTimer)
    this.completions.textContent = matches.join(' | ')
    this.completions.hidden = !matches.length
    this.completionTimer = setTimeout(() => this.hideCompletions(), 5000)
  }

  hideCompletions () {
    clearTimeout(this.completionTimer)
    this.completions.hidden = true
  }
}
