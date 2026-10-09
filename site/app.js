import challengesJson from './challenges.json'
import { filterChallenges, siteConfig } from './config.js'
import { runCommand, fetchSolutions } from './api.js'
import { loadProgress, saveProgress } from './storage.js'
import { startRouter, navigate } from './router.js'
import { CommandInput } from './command-input.js'
import { UI } from './ui.js'

export function startApp () {
  const config = siteConfig(window.location.hostname)
  const challenges = filterChallenges(challengesJson, config.flavor)
  const progress = loadProgress()
  const ui = new UI(config, challenges)
  let current
  let solutionsRequest
  const nextChallenge = () => challenges.find(challenge => !progress.has(challenge.slug)) || challenges[0]

  const input = new CommandInput({
    onSubmit: submit,
    onClear: () => { ui.hideInfo(); ui.showOutput() },
    candidates: () => [...config.commands, ...(current.completions || [])]
  })

  async function loadSolutions (challenge) {
    solutionsRequest?.abort()
    const controller = new AbortController()
    solutionsRequest = controller
    ui.showSolutions([], 'Loading solutions…')
    try {
      const commands = await fetchSolutions(challenge.slug, controller.signal)
      if (!controller.signal.aborted && current === challenge) {
        ui.showSolutions(commands, commands.length ? '' : 'No solutions yet for this challenge')
      }
    } catch (error) {
      if (error.name !== 'AbortError' && current === challenge) {
        ui.showSolutions([], 'Unable to fetch solutions')
      }
    }
  }

  function selectChallenge (challenge) {
    if (current === challenge) return
    current = challenge
    ui.hideInfo()
    ui.showOutput()
    input.hideCompletions()
    ui.renderChallenge(current, progress)
    loadSolutions(current)
  }

  async function submit (command) {
    const submitted = current
    navigate(submitted.slug)
    ui.hideInfo()
    ui.showOutput()
    document.getElementById('chck1').checked = false
    if (/tail\s+-[Ff]/.test(command)) {
      ui.showInfo('tail -f will wait for additional data to be appended to the file, try removing the -f option', 'incorrect', submitted)
      input.focus()
      return
    }
    input.setBusy(true)
    try {
      const response = await runCommand(command, submitted)
      input.setExitCode(response.ExitCode)
      if (response.Correct) {
        progress.add(submitted.slug)
        saveProgress(progress)
        if (current === submitted) {
          const next = nextChallenge()
          selectChallenge(next)
          navigate(next.slug)
          // Refresh badges even when completing the final challenge leaves the route unchanged.
          ui.renderChallenge(current, progress)
          const won = challenges.every(challenge => progress.has(challenge.slug))
          ui.showInfo(won
            ? 'Correct! You have completed all of the challenges, but feel free to keep on going!'
            : 'Correct! You have a new challenge!', 'correct', submitted)
          ui.showOutput(response.Output)
        } else {
          ui.renderChallenge(current, progress)
        }
      } else if (current === submitted) {
        ui.showOutput(response.Output)
        ui.showInfo(response.Error ? response.Error + ' - try again' : 'Incorrect answer, try again', 'incorrect', submitted)
      }
    } catch (error) {
      input.setExitCode(null)
      if (current === submitted) ui.showInfo(error.message || 'Unknown Error :(', 'error', submitted)
    } finally {
      input.setBusy(false)
    }
  }

  startRouter(challenges, nextChallenge, selectChallenge)
  document.getElementById('badges').addEventListener('click', event => {
    if (event.target.closest('a')) input.focus()
  })
  if (window.matchMedia('(pointer: fine)').matches) input.focus()
}
