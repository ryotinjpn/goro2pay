import { test } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const SHOT_DIR = path.resolve(__dirname, '../.shots');
fs.mkdirSync(SHOT_DIR, { recursive: true });

test.use({ viewport: { width: 360, height: 800 } });

test('Win verdict fits on narrow mobile width', async ({ page }) => {
  await page.goto('/main');
  await page.waitForTimeout(200);
  await page
    .locator('button:has-text("めんどくさい"), button:has-text("押す。")')
    .first()
    .click();
  await page.waitForTimeout(2900);
  await page.screenshot({
    path: path.join(SHOT_DIR, 'win-narrow.png'),
    fullPage: false,
  });
});
