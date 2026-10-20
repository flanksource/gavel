import { defineConfig } from '@playwright/test';

// Browser specs drive a real `gavel serve` that a harness (hack/*_e2e.go)
// starts and names through env. Inside a gavel fixture with an `http_proxy` trace,
// HTTP_PROXY is the recording proxy: Chromium ignores the env variable, so it
// is passed explicitly, and `<-loopback>` stops Chromium bypassing it for the
// loopback dashboard.
const recordingProxy = process.env.HTTP_PROXY ?? process.env.http_proxy;

export default defineConfig({
  testDir: 'e2e',
  fullyParallel: false,
  workers: 1,
  retries: 0,
  reporter: 'line',
  timeout: 120_000,
  use: {
    headless: true,
    trace: 'retain-on-failure',
    ...(recordingProxy ? { proxy: { server: recordingProxy, bypass: '<-loopback>' } } : {}),
  },
});
