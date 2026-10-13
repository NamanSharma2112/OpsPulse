import { defineConfig, devices } from '@playwright/test';
import dotenv from 'dotenv';

// Load ABLESPACE_EMAIL / ABLESPACE_PASSWORD / BASE_URL from .env (git-ignored).
dotenv.config();

export default defineConfig({
  testDir: './tests',
  timeout: 60_000,
  expect: { timeout: 10_000 },

  // This runs against a live staging app with one real account, so keep the
  // traffic small and the order predictable: one worker, no automatic retries.
  fullyParallel: false,
  workers: 1,
  retries: 0,

  reporter: [['list'], ['html', { open: 'never' }]],

  use: {
    baseURL: process.env.BASE_URL || 'https://staging.ablespace.io',
    // Keep debugging material only when a test fails.
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
});
