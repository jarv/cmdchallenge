import hljs from 'highlight.js/lib/core'
import bash from 'highlight.js/lib/languages/bash'
import { Parser, HtmlRenderer } from 'commonmark'

hljs.registerLanguage('bash', bash)
const reader = new Parser()
const writer = new HtmlRenderer()
const markdown = text => writer.render(reader.parse(text))
const element = id => document.getElementById(id)

const emojiGroups = {
  error: ['emojis/1F63F.png'],
  incorrect: ['emojis/E282.png', 'emojis/1F645-200D-2640-FE0F.png', 'emojis/1F645-200D-2642-FE0F.png', 'emojis/1F940.png'],
  beginner: ['emojis/1F471-200D-2640-FE0F.png', 'emojis/1F471-200D-2642-FE0F.png'],
  intermediate: ['emojis/1F9D4.png', 'emojis/1F468-200D-1F9B1.png', 'emojis/1F468-200D-1F9B0.png', 'emojis/1F468-200D-1F33E.png', 'emojis/1F468-200D-1F52C.png', 'emojis/1F468-200D-1F373.png', 'emojis/1F468-200D-1F393.png'],
  advanced: ['emojis/1F478.png', 'emojis/1F482.png', 'emojis/1F9DD.png', 'emojis/1F9DD-200D-2640-FE0F.png', 'emojis/1F680.png'],
  oops: ['emojis/1F600.png', 'emojis/1F604.png', 'emojis/1F970.png', 'emojis/1F60D.png', 'emojis/1F929.png'],
  '12days': ['emojis/1F936.png', 'emojis/1F385.png', 'emojis/1F36D.png', 'emojis/2603.png']
}

function image (filename, alt = '') {
  const img = document.createElement('img')
  img.src = '/img/' + filename
  img.alt = alt
  return img
}

export class UI {
  constructor (config, challenges) {
    this.config = config
    this.challenges = challenges
    this.emojiIndexes = {}
    document.title = config.flavor === '12days'
      ? '🎄 Twelve Days of Shell 🎄'
      : config.flavor === 'oops' ? 'Oops, I deleted my bin/ dir :(' : 'Command Challenge!'
    element('header-img').replaceChildren(image(config.image))
    element('header-text').textContent = config.title
    if (config.flavor === '12days') {
      const fonts = document.createElement('link')
      fonts.rel = 'stylesheet'
      fonts.href = '/fonts/fonts.css'
      document.head.append(fonts)
      element('header-text').classList.add('snowburst')
    }
  }

  renderChallenge (challenge, progress) {
    element('challenge-desc').querySelector('.img-container').replaceChildren(image(challenge.emoji + '.png', challenge.disp_title))
    // Descriptions and learn content are repository-authored Markdown, including HTML.
    element('challenge-desc').querySelector('.desc-container').innerHTML = markdown(challenge.description)
    element('learn-box').hidden = !challenge.learn
    element('learn').innerHTML = challenge.learn ? markdown(challenge.learn) : ''
    element('chck2').checked = !!challenge.disp_learn
    this.renderBadges(challenge, progress)
    this.renderWin(progress)
  }

  renderBadges (current, progress) {
    const next = this.challenges.find(challenge => !progress.has(challenge.slug))
    const badges = this.challenges.filter(challenge => progress.has(challenge.slug) || challenge === next)
    element('badges').replaceChildren(...badges.map(challenge => {
      const wrapper = document.createElement('div')
      const active = challenge === current
      wrapper.className = 'img-container ' + (active ? 'active-badge' : 'inactive-badge')
      const link = document.createElement('a')
      link.href = '#/' + challenge.slug
      link.id = 'badge_' + challenge.slug
      link.title = challenge.disp_title
      if (active) link.setAttribute('aria-current', 'page')
      const img = image(challenge.emoji + '.png', challenge.disp_title)
      img.className = 'badge'
      link.append(img)
      wrapper.append(link)
      return wrapper
    }))
  }

  renderWin (progress) {
    const won = document.querySelector('.title .won')
    won.hidden = this.challenges.some(challenge => !progress.has(challenge.slug))
    if (won.hidden) return
    won.textContent = this.config.flavor === '12days'
      ? '🎄 Congrats, you completed all 12 days! 🎄 '
      : '🎉 Congrats, you completed the challenge! 🎉 '
    if (this.config.flavor !== 'cmdchallenge') {
      const link = document.createElement('a')
      link.href = this.config.home
      link.textContent = 'Try even more challenges!'
      won.append(link)
    }
  }

  hideInfo () {
    element('info-box').hidden = true
  }

  showInfo (message, status, challenge) {
    const box = element('info-box')
    box.querySelector('.text').textContent = message
    box.querySelector('.gradient').className = 'gradient ' + status
    const index = this.challenges.indexOf(challenge)
    const group = status !== 'correct'
      ? status
      : this.config.flavor !== 'cmdchallenge'
        ? this.config.flavor
        : index < 4 ? 'beginner' : index < 20 ? 'intermediate' : 'advanced'
    const items = emojiGroups[group]
    const emojiIndex = this.emojiIndexes[group] || 0
    this.emojiIndexes[group] = emojiIndex + 1
    box.querySelector('.img').replaceChildren(image(items[emojiIndex % items.length], status))
    box.hidden = false
  }

  showOutput (output = '') {
    const container = element('challenge-output')
    container.replaceChildren(...output.split('\n').map(line => {
      const span = document.createElement('span')
      span.textContent = line
      return span
    }))
    container.hidden = !output
  }

  showSolutions (commands, status = '') {
    const code = element('solutions')
    code.textContent = commands.join('\n')
    code.hidden = !commands.length
    delete code.dataset.highlighted
    if (commands.length) hljs.highlightElement(code)
    element('solutions-status').textContent = status
  }
}
