export function navigate (slug) {
  window.location.hash = '/' + slug
}

export function startRouter (challenges, fallback, onChange) {
  const update = () => {
    const slug = window.location.hash.replace(/^#\//, '')
    onChange(challenges.find(challenge => challenge.slug === slug) || fallback())
  }
  window.addEventListener('hashchange', update)
  update()
}
