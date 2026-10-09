import { readFile } from 'node:fs/promises'
import { test, expect } from '@playwright/test'

const challenges = JSON.parse(await readFile(new URL('../challenges.json', import.meta.url), 'utf8'))
const main = challenges.filter(challenge => !challenge.tags)
let pageErrors

test.beforeEach(async ({ page }) => {
  pageErrors = []
  page.on('pageerror', error => pageErrors.push(error.message))
  await page.route('**/s.cmdchallenge.com/**', route => route.fulfill({ body: '', contentType: 'text/javascript' }))
  await page.route('**/c/s?*', route => route.fulfill({ json: { cmds: [] } }))
})

test.afterEach(() => {
  expect(pageErrors).toEqual([])
})

test('runs a command, prevents duplicate submissions, advances, and persists progress', async ({ page }) => {
  let release
  const ready = new Promise(resolve => { release = resolve })
  let requests = 0
  await page.route('**/c/r', async route => {
    requests++
    const form = new URLSearchParams(route.request().postData())
    expect(form.get('slug')).toBe(main[0].slug)
    expect(Buffer.from(form.get('cmd'), 'base64').toString('utf8')).toBe('echo hello world')
    await ready
    await route.fulfill({ json: { Correct: true, ExitCode: 0, Output: 'hello world' } })
  })
  await page.goto('/')
  const input = page.locator('#command-input')
  await input.fill('  echo hello world  ')
  await page.getByRole('button', { name: 'Run', exact: true }).click()
  await expect(page.locator('#term-spinner')).toBeVisible()
  await expect(page.locator('#command-form')).toBeHidden()
  await expect(page.locator('.command-prompt')).toBeHidden()
  await expect(page.locator('.command-cursor')).toBeHidden()
  await expect(page.getByRole('button', { name: 'Run', exact: true, includeHidden: true })).toBeHidden()
  await expect(input).toHaveAttribute('readonly', '')
  await page.locator('#command-form').evaluate(form => { form.requestSubmit(); form.requestSubmit() })
  release()
  await expect(page).toHaveURL(new RegExp('#/' + main[1].slug + '$'))
  await expect(input).toBeFocused()
  await expect(page.locator('#command-form')).toBeVisible()
  await expect(page.locator('#term-spinner')).toBeHidden()
  await expect(page.locator('#info-box')).toContainText('Correct! You have a new challenge!')
  await expect(page.locator('#challenge-output')).toHaveText('hello world')
  expect(requests).toBe(1)
  await page.reload()
  await expect(page.locator('#badge_' + main[0].slug)).toBeVisible()
  expect(await page.evaluate(() => JSON.parse(localStorage.getItem('correct_answers')))).toContain(main[0].slug)
})

test('green Run button is replaced by a rotating, bouncing line while running', async ({ page }) => {
  let release
  const ready = new Promise(resolve => { release = resolve })
  await page.route('**/c/r', async route => {
    await ready
    await route.fulfill({ status: 500, body: 'Try again' })
  })
  await page.goto('/')
  const button = page.getByRole('button', { name: 'Run', exact: true })
  await expect(button).toHaveCSS('color', 'rgb(3, 242, 0)')
  await expect(button).toHaveCSS('border-top-color', 'rgb(3, 242, 0)')
  await expect(button).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
  await page.locator('#command-input').fill('ls')
  await button.click()
  const spinner = page.locator('.command-spinner')
  await expect(spinner).toBeVisible()
  await expect(page.locator('#term-challenge')).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
  const track = await page.locator('#term-spinner').boundingBox()
  const positions = await spinner.evaluate(element => {
    const animation = element.getAnimations().find(animation => animation.animationName === 'command-spinner-travel')
    animation.pause()
    return [0, 2000, 4000].map(time => {
      animation.currentTime = time
      return element.getBoundingClientRect().left
    })
  })
  expect(positions[0]).toBeCloseTo(track.x, 0)
  expect(positions[1]).toBeGreaterThan(track.x + track.width * 0.9)
  expect(positions[2]).toBeCloseTo(track.x, 0)
  const frames = await spinner.evaluate(element => {
    const animation = element.getAnimations({ subtree: true }).find(animation => animation.animationName === 'command-spinner-rotate')
    animation.pause()
    return [50, 150, 250, 350].map(time => {
      animation.currentTime = time
      return JSON.parse(getComputedStyle(element, '::before').content)
    })
  })
  expect(frames).toEqual(['|', '/', '-', '\\'])
  release()
  await expect(page.locator('#term-spinner')).toBeHidden()
  await expect(page.locator('#command-form')).toBeVisible()
  await expect(page.locator('#command-input')).toBeFocused()
  await expect(page.locator('#term-challenge')).toHaveCSS('background-color', 'rgb(26, 26, 26)')
  await expect(page.locator('.command-cursor')).toBeVisible()
})

test('renders output and server messages literally and restores input after errors', async ({ page }) => {
  await page.route('**/c/r', route => route.fulfill({
    json: { Correct: false, ExitCode: 1, Output: '<img src=x onerror="alert(1)">\nsecond line', Error: '<b>try again</b>' }
  }))
  await page.goto('/')
  const input = page.locator('#command-input')
  await input.fill('false')
  await input.press('Enter')
  await expect(page.locator('#command-exit-code')).toHaveText('1')
  await expect(page.locator('#challenge-output span')).toHaveCount(2)
  await expect(page.locator('#challenge-output img')).toHaveCount(0)
  await expect(page.locator('#info-box .text')).toHaveText('<b>try again</b> - try again')
  await expect(input).toBeFocused()

  await page.route('**/c/r', route => route.fulfill({ status: 429, body: '<b>Slow down!</b>' }))
  await input.fill('ls')
  await input.press('Enter')
  await expect(page.locator('#info-box .text')).toHaveText('<b>Slow down!</b>')
  await expect(page.locator('#command-exit-code')).toHaveText('☠️')
  await expect(input).not.toHaveAttribute('readonly', '')
  await expect(page.locator('#term-spinner')).toBeHidden()
})

test('recovers from network failures and malformed JSON', async ({ page }) => {
  await page.route('**/c/r', route => route.abort('failed'))
  await page.goto('/')
  const input = page.locator('#command-input')
  await input.fill('ls')
  await input.press('Enter')
  await expect(page.locator('#info-box .gradient')).toHaveClass('gradient error')
  await expect(input).not.toHaveAttribute('readonly', '')
  await page.route('**/c/r', route => route.fulfill({ body: 'not JSON' }))
  await input.fill('pwd')
  await input.press('Enter')
  await expect(page.locator('#term-spinner')).toBeHidden()
  await expect(input).toBeFocused()
})

test('history restores drafts, migrates old commands, and survives reloads', async ({ page }) => {
  await page.addInitScript(() => localStorage.setItem('cmdchallenge_0_commands', '["echo old command"]'))
  await page.route('**/c/r', route => route.fulfill({ json: { Correct: false, ExitCode: 0 } }))
  await page.goto('/')
  const input = page.locator('#command-input')
  await input.fill('draft')
  await input.press('ArrowUp')
  await expect(input).toHaveValue('echo old command')
  await input.press('ArrowDown')
  await expect(input).toHaveValue('draft')
  await input.fill('echo new command')
  await input.press('Enter')
  await expect(page.locator('#info-box')).toBeVisible()
  await page.reload()
  await input.press('ArrowUp')
  await expect(input).toHaveValue('echo new command')
})

test('completion, clearing, IME, and keyboard focus navigation work', async ({ page }) => {
  await page.goto('/')
  const input = page.locator('#command-input')
  await input.fill('gr')
  await input.press('Tab')
  await expect(input).toHaveValue('grep')
  await input.press('Tab')
  await expect(page.locator('#completions')).toHaveText('grep')
  await input.press('Escape')
  await expect(page.locator('#completions')).toBeHidden()
  await input.fill('echo composing')
  await input.dispatchEvent('keydown', { key: 'Enter', isComposing: true })
  await expect(input).toHaveValue('echo composing')
  await input.fill('')
  await input.press('Tab')
  await expect(page.getByRole('button', { name: 'Run', exact: true })).toBeFocused()
  await input.focus()
  await input.press('Shift+Tab')
  await expect(input).not.toBeFocused()
  await input.focus()
  await input.press('Enter')
  await expect(page.locator('#command-exit-code')).toHaveText('0')
})

test('clear is local, resets the terminal, and leaves history and progress intact', async ({ page }) => {
  let requests = 0
  await page.route('**/c/r', route => {
    requests++
    return route.fulfill({ json: { Correct: false, ExitCode: 1, Output: 'previous output' } })
  })
  await page.goto('/')
  const input = page.locator('#command-input')
  await input.fill('false')
  await input.press('Enter')
  await expect(page.locator('#challenge-output')).toHaveText('previous output')
  await input.fill('cle')
  await input.press('Tab')
  await expect(input).toHaveValue('clear')
  await input.press('Enter')
  await expect(input).toHaveValue('')
  await expect(input).toBeFocused()
  await expect(page.locator('#challenge-output')).toBeHidden()
  await expect(page.locator('#info-box')).toBeHidden()
  await expect(page.locator('#command-exit-code')).toHaveText('0')
  await expect(page.locator('#term-spinner')).toBeHidden()
  expect(requests).toBe(1)
  await input.press('ArrowUp')
  await expect(input).toHaveValue('false')
  expect(await page.evaluate(() => localStorage.getItem('correct_answers'))).toBeNull()
})

test('Ctrl+A and Ctrl+E move to logical line boundaries without selecting or changing text', async ({ page }) => {
  await page.goto('/')
  const input = page.locator('#command-input')
  await input.fill('echo first\necho second\necho third')
  await input.evaluate(element => element.setSelectionRange(14, 14))
  await input.press('Control+a')
  expect(await input.evaluate(element => [element.selectionStart, element.selectionEnd])).toEqual([11, 11])
  await expect(page.locator('.command-cursor')).toHaveText('e')
  await input.press('Control+e')
  expect(await input.evaluate(element => [element.selectionStart, element.selectionEnd])).toEqual([22, 22])
  await input.evaluate(element => element.setSelectionRange(0, 0))
  await input.press('Control+a')
  expect(await input.evaluate(element => element.selectionStart)).toBe(0)
  await input.press('Control+e')
  expect(await input.evaluate(element => element.selectionStart)).toBe(10)
  await expect(input).toHaveValue('echo first\necho second\necho third')

  // A wrapped logical line can span beyond the textarea viewport.
  await input.fill('x'.repeat(5000))
  await input.press('Control+a')
  expect(await input.evaluate(element => [element.selectionStart, element.selectionEnd])).toEqual([0, 0])
  expect(await input.evaluate(element => element.scrollTop)).toBeLessThan(30)
  await input.press('Control+e')
  expect(await input.evaluate(element => [element.selectionStart, element.selectionEnd])).toEqual([5000, 5000])
  const field = await input.boundingBox()
  const cursor = await page.locator('.command-cursor').boundingBox()
  expect(cursor.y).toBeGreaterThanOrEqual(field.y)
  expect(cursor.y + cursor.height).toBeLessThanOrEqual(field.y + field.height)
})

test('Unicode and multiline commands retain their bytes in the form request', async ({ page }) => {
  let command
  await page.route('**/c/r', route => {
    command = Buffer.from(new URLSearchParams(route.request().postData()).get('cmd'), 'base64').toString('utf8')
    return route.fulfill({ json: { Correct: false, ExitCode: 0 } })
  })
  await page.goto('/')
  const input = page.locator('#command-input')
  await input.fill('echo café 🐚\necho second line')
  await input.press('Enter')
  await expect(page.locator('#info-box')).toBeVisible()
  expect(command).toBe('echo café 🐚\necho second line')
})

test('input has no outline and its blinking green block follows the caret', async ({ page }) => {
  await page.goto('/')
  const input = page.locator('#command-input')
  const cursor = page.locator('.command-cursor')
  await input.focus()
  await expect(input).toHaveCSS('outline-style', 'none')
  await expect(cursor).toBeVisible()
  await expect(cursor).toHaveCSS('background-color', 'rgb(3, 242, 0)')
  await expect(cursor).toHaveCSS('animation-name', 'command-cursor-blink')
  await cursor.evaluate(element => {
    const animation = element.getAnimations()[0]
    animation.pause()
    animation.currentTime = 750
  })
  await expect(cursor).toHaveCSS('opacity', '0')
  await cursor.evaluate(element => { element.getAnimations()[0].currentTime = 250 })
  await expect(cursor).toHaveCSS('opacity', '1')
  const origin = await input.boundingBox()
  await input.fill('abcdef')
  await input.press('ArrowLeft')
  await input.press('ArrowLeft')
  await expect(cursor).toHaveText('e')
  const block = await cursor.boundingBox()
  expect(block.x - origin.x).toBeCloseTo(block.width * 4, 0)
  await input.press('Shift+ArrowLeft')
  await expect(cursor).toBeHidden()
  await page.getByRole('button', { name: 'Run', exact: true }).focus()
  await expect(cursor).toBeHidden()
})

test('block cursor tracks multiline wrapping and scroll without changing editing', async ({ page }) => {
  await page.goto('/')
  await page.emulateMedia({ reducedMotion: 'reduce' })
  const input = page.locator('#command-input')
  const cursor = page.locator('.command-cursor')
  await input.fill('first line\nsecond line')
  await expect(cursor).toBeVisible()
  await expect(cursor).toHaveCSS('animation-name', 'none')
  const origin = await input.boundingBox()
  const block = await cursor.boundingBox()
  expect(block.y - origin.y).toBeGreaterThan(30)
  expect(block.x - origin.x).toBeCloseTo(block.width * 11, 0)
  await input.fill('x'.repeat(120))
  const wrapped = await cursor.boundingBox()
  expect(wrapped.y - origin.y).toBeGreaterThan(30)
  await input.fill('x'.repeat(5000))
  await input.evaluate(element => {
    element.scrollTop = element.scrollHeight
    element.dispatchEvent(new Event('scroll'))
  })
  const scrolledInput = await input.boundingBox()
  const scrolledCursor = await cursor.boundingBox()
  expect(scrolledCursor.y).toBeGreaterThanOrEqual(scrolledInput.y)
  expect(scrolledCursor.y + scrolledCursor.height).toBeLessThanOrEqual(scrolledInput.y + scrolledInput.height)
  await input.dispatchEvent('compositionstart')
  await expect(cursor).toBeHidden()
  await input.dispatchEvent('compositionend')
  await expect(cursor).toBeVisible()
})

test('a response completes its submitted challenge even after navigation', async ({ page }) => {
  let release
  const ready = new Promise(resolve => { release = resolve })
  await page.route('**/c/r', async route => {
    await ready
    await route.fulfill({ json: { Correct: true, ExitCode: 0, Output: 'old response' } })
  })
  await page.goto('/')
  await page.locator('#command-input').fill('echo hello world')
  await page.locator('#command-input').press('Enter')
  await expect(page.locator('#term-spinner')).toBeVisible()
  await page.evaluate(slug => { window.location.hash = '/' + slug }, main[2].slug)
  await expect(page.locator('#challenge-desc .desc-container')).toContainText(main[2].description.split('\n')[0])
  release()
  await expect(page.locator('#term-spinner')).toBeHidden()
  expect(await page.evaluate(() => JSON.parse(localStorage.getItem('correct_answers')))).toEqual([main[0].slug])
  await expect(page).toHaveURL(new RegExp('#/' + main[2].slug + '$'))
  await expect(page.locator('#challenge-output')).toBeHidden()
})

test('solutions ignore stale responses and display empty and failed states', async ({ page }) => {
  let release
  const ready = new Promise(resolve => { release = resolve })
  await page.route('**/c/s?*', async route => {
    const slug = new URL(route.request().url()).searchParams.get('slug')
    if (slug === main[0].slug) await ready
    await route.fulfill({ json: { cmds: [slug === main[0].slug ? 'old solution' : 'new solution'] } }).catch(() => {})
  })
  await page.goto('/')
  await page.evaluate(slug => { window.location.hash = '/' + slug }, main[1].slug)
  await page.getByText('View Solutions', { exact: true }).click()
  await expect(page.locator('#solutions')).toHaveText('new solution')
  release()
  await expect(page.locator('#solutions')).toHaveText('new solution')
  await page.route('**/c/s?*', route => route.fulfill({ json: { cmds: [] } }))
  await page.goto('/#/' + main[2].slug)
  await expect(page.locator('#solutions-status')).toHaveText('No solutions yet for this challenge')
  await page.route('**/c/s?*', route => route.fulfill({ status: 500 }))
  await page.goto('/#/' + main[3].slug)
  await expect(page.locator('#solutions-status')).toHaveText('Unable to fetch solutions')
})

test('deep links, back/forward, invalid routes, and Learn visibility work', async ({ page }) => {
  await page.goto('/#/' + main[1].slug)
  await expect(page.locator('#challenge-desc .desc-container')).toBeVisible()
  await page.evaluate(slug => { window.location.hash = '/' + slug }, main[2].slug)
  await expect(page).toHaveURL(new RegExp(main[2].slug + '$'))
  await page.goBack()
  await expect(page).toHaveURL(new RegExp(main[1].slug + '$'))
  await page.goForward()
  await expect(page).toHaveURL(new RegExp(main[2].slug + '$'))
  await page.goto('/#/missing')
  await expect(page.locator('#badge_' + main[0].slug)).toHaveAttribute('aria-current', 'page')
  await expect(page.locator('#learn-box')).toBeHidden()
})

test('completing the final challenge displays the win state', async ({ page }) => {
  await page.addInitScript(slugs => localStorage.setItem('correct_answers', JSON.stringify(slugs)), main.slice(0, -1).map(challenge => challenge.slug))
  await page.route('**/c/r', route => route.fulfill({ json: { Correct: true, ExitCode: 0 } }))
  await page.goto('/')
  await page.locator('#command-input').fill('last answer')
  await page.locator('#command-input').press('Enter')
  await expect(page.locator('.won')).toBeVisible()
  await expect(page.locator('#info-box')).toContainText('completed all of the challenges')
})

for (const flavor of ['oops', '12days']) {
  test(`${flavor} flavor loads its challenges and assets without legacy globals`, async ({ page }) => {
    const flavorChallenges = challenges.filter(challenge => challenge.tags?.includes(flavor))
    await page.goto(`http://${flavor}.localhost:4191/`)
    await expect(page.locator('#badge_' + flavorChallenges[0].slug)).toBeVisible()
    await expect(page.locator('#header-text')).toContainText(flavor === 'oops' ? 'Oops' : 'Twelve Days')
    if (flavorChallenges[0].learn) await expect(page.locator('#learn-box')).toBeVisible()
    else await expect(page.locator('#learn-box')).toBeHidden()
    expect(await page.evaluate(() => typeof window.jQuery)).toBe('undefined')
    if (flavor === '12days') {
      await expect(page.locator('#chck2')).toBeChecked()
      await page.evaluate(() => document.fonts.ready)
      expect(await page.evaluate(() => document.fonts.check('16px "Snowburst One"'))).toBe(true)
    }
  })
}
