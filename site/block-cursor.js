// Mirror native textarea layout so the block follows wrapping, scrolling, and
// caret movement without replacing the browser's editing or selection behavior.
export class BlockCursor {
  constructor (input) {
    this.input = input
    this.mirror = input.parentElement.querySelector('.command-mirror')
    this.cursor = input.parentElement.querySelector('.command-cursor')
    this.composing = false
    for (const event of ['input', 'keyup', 'click', 'select', 'scroll', 'focus', 'blur']) {
      input.addEventListener(event, () => this.update())
    }
    document.addEventListener('selectionchange', () => this.update())
    input.addEventListener('compositionstart', () => {
      this.composing = true
      this.update()
    })
    input.addEventListener('compositionend', () => {
      this.composing = false
      this.update()
    })
    // Let the native caret guide IME composition while the block is hidden.
    input.addEventListener('compositionstart', () => input.classList.add('composing'))
    input.addEventListener('compositionend', () => input.classList.remove('composing'))
  }

  update ({ reveal = false } = {}) {
    const input = this.input
    const visible = document.activeElement === input && !input.readOnly &&
      !this.composing && input.selectionStart === input.selectionEnd
    this.cursor.hidden = !visible
    if (!visible) return

    this.mirror.style.width = `${input.clientWidth}px`
    // A final zero-width character gives the caret a measurable line box even
    // when the command is empty or ends with a newline.
    this.mirror.textContent = input.value + '\u200b'
    const range = document.createRange()
    range.setStart(this.mirror.firstChild, input.selectionStart)
    range.collapse(true)
    const rect = range.getClientRects()[0]
    if (!rect) {
      this.cursor.hidden = true
      return
    }
    const origin = this.mirror.getBoundingClientRect()
    if (reveal) {
      const top = rect.top - origin.top
      if (top < input.scrollTop) input.scrollTop = top
      else if (top + rect.height > input.scrollTop + input.clientHeight) {
        input.scrollTop = top + rect.height - input.clientHeight
      }
    }
    const fontSize = parseFloat(getComputedStyle(input).fontSize)
    this.cursor.style.left = `${rect.left - origin.left - input.scrollLeft}px`
    this.cursor.style.top = `${rect.top - origin.top - input.scrollTop + (rect.height - fontSize) / 2}px`
    const character = String.fromCodePoint(input.value.codePointAt(input.selectionStart) || 32)
    this.cursor.textContent = /\s/.test(character) ? '' : character
  }
}
