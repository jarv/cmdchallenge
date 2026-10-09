function encodeCommand (command) {
  const bytes = new TextEncoder().encode(command)
  return btoa(Array.from(bytes, byte => String.fromCharCode(byte)).join(''))
}

async function request (url, options) {
  const response = await fetch(url, options)
  if (!response.ok) {
    throw new Error((await response.text()).trim() || `Request failed (${response.status})`)
  }
  return response.json()
}

export async function runCommand (command, challenge) {
  const response = await request('/c/r', {
    method: 'POST',
    body: new URLSearchParams({
      cmd: encodeCommand(command),
      slug: challenge.slug,
      version: challenge.version,
      img: challenge.img || 'cmd'
    })
  })
  if (typeof response.Correct !== 'boolean' || !Number.isInteger(response.ExitCode)) {
    throw new Error('Invalid response from command runner')
  }
  return response
}

export async function fetchSolutions (slug, signal) {
  const response = await request('/c/s?' + new URLSearchParams({ slug }), { signal })
  if (!Array.isArray(response.cmds) || !response.cmds.every(cmd => typeof cmd === 'string')) {
    throw new Error('Invalid solutions response')
  }
  return response.cmds
}
