import { test } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const SHOT_DIR = path.resolve(__dirname, '../.shots');
fs.mkdirSync(SHOT_DIR, { recursive: true });

test('Capture win confirmation choreography', async ({ page }) => {
  await page.goto('/main');
  await page.waitForTimeout(200);
  await page
    .locator('button:has-text("めんどくさい"), button:has-text("押す。")')
    .first()
    .click();

  // Wait until reel stops (placeOrder = 2.8s)
  await page.waitForTimeout(2700);

  // capture frames around stop moment
  for (let i = 0; i < 12; i++) {
    await page.waitForTimeout(80);
    await page.screenshot({
      path: path.join(SHOT_DIR, `win-${String(i).padStart(2, '0')}.png`),
    });
  }
});
