import { defineConfig } from 'vite'

export default defineConfig({
  build: {
    rollupOptions: {
      external: [
        'css/jquery.terminal.min.css'
      ]
    }
  }
})
