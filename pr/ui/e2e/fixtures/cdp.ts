import { test as base, expect } from '@playwright/test';

// Specs run by a gavel fixture `playwright` step import { test, expect } from
// here instead of '@playwright/test'. When the fixture's `browser` step has
// exported BROWSER_CDP_URL, `browser`/`context`/`page` attach to that Chrome
// over CDP and reuse its default context, so every step shares one profile
// (cookies, storage) and one process. Without it, the stock fixtures launch a
// browser as usual.
//
// The shared Chrome belongs to the `browser` step: nothing here closes it or
// its default context. Only the page a test opened is closed.
const cdpURL = process.env.BROWSER_CDP_URL;

export const test = cdpURL
  ? base.extend({
      browser: [
        async ({ playwright }, use) => {
          await use(await playwright.chromium.connectOverCDP(cdpURL));
        },
        { scope: 'worker' },
      ],
      context: async ({ browser }, use) => {
        const [context] = browser.contexts();
        if (!context) throw new Error(`Chrome at ${cdpURL} exposes no default browser context`);
        await use(context);
      },
      page: async ({ context }, use) => {
        const page = await context.newPage();
        await use(page);
        await page.close();
      },
    })
  : base;

export { expect };
