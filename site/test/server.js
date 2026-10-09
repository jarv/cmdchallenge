import http from 'node:http'
import { readFile } from 'node:fs/promises'
import path from 'node:path'

const root = path.resolve('dist')
const types = { '.html': 'text/html', '.js': 'text/javascript', '.css': 'text/css', '.png': 'image/png', '.svg': 'image/svg+xml', '.woff2': 'font/woff2' }
http.createServer(async (request, response) => {
  const pathname = decodeURIComponent(new URL(request.url, 'http://localhost').pathname)
  const filename = path.resolve(root, '.' + (pathname === '/' ? '/index.html' : pathname))
  if (!filename.startsWith(root + path.sep)) {
    response.writeHead(403).end()
    return
  }
  try {
    const content = await readFile(filename)
    response.writeHead(200, { 'Content-Type': types[path.extname(filename)] || 'application/octet-stream' })
    response.end(content)
  } catch {
    response.writeHead(404).end()
  }
}).listen(4191, '127.0.0.1')
