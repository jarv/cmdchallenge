const COMMANDS = ['find', 'echo', 'awk', 'sed', 'perl', 'wc', 'grep', 'cat', 'sort', 'cut', 'ls', 'tac', 'jq', 'paste', 'tr', 'rm', 'tail', 'comm', 'egrep']

export function siteConfig (hostname) {
  const parts = hostname.split('.')
  const flavor = ['oops', '12days'].includes(parts[0]) ? parts[0] : 'cmdchallenge'
  const base = parts.filter(part => !['oops', '12days'].includes(part)).join('.')
  return {
    flavor,
    home: '//' + base,
    commands: flavor === 'oops' ? ['echo', 'read'] : COMMANDS,
    title: flavor === 'oops'
      ? 'Oops I deleted my bin/ dir :('
      : flavor === '12days' ? 'Twelve Days of Shell' : 'Command Challenge',
    image: flavor === 'oops' ? 'emojis/1F92D.png' : flavor === '12days' ? 'emojis/1F384.png' : 'cmdchallenge-round.png'
  }
}

export function filterChallenges (challenges, flavor) {
  return challenges.filter(challenge => flavor === 'cmdchallenge'
    ? !challenge.tags
    : (challenge.tags || []).includes(flavor))
}
