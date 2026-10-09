import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './test',
  testMatch: '**/*.spec.js',
  use: { baseURL: 'http://localhost:4191', trace: 'retain-on-failure' },
  projects: [
    { name: 'desktop', use: { ...devices['Desktop Chrome'] } },
    { name: 'mobile', use: { ...devices['Pixel 5'] } }
  ],
  webServer: {
    command: 'node test/server.js',
    url: 'http://localhost:4191',
    reuseExistingServer: !process.env.CI
  }
})
