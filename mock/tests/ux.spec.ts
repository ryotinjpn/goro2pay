import { test, expect } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const SHOT_DIR = path.resolve(__dirname, '../.shots');
fs.mkdirSync(SHOT_DIR, { recursive: true });


test('Landing screen renders', async ({ page }) => {
  await page.goto('/');
  await page.screenshot({ path: path.join(SHOT_DIR, '01-landing.png'), fullPage: true });

  await expect(page.getByText('考えるな。')).toBeVisible();
  await expect(page.getByText('押せ。')).toBeVisible();
});

test('Landing demo button triggers slot once', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: '押す。' }).click();
  await page.waitForTimeout(400);
  await page.screenshot({ path: path.join(SHOT_DIR, '02-landing-slot.png'), fullPage: true });
  await page.waitForTimeout(1500);
  await page.screenshot({ path: path.join(SHOT_DIR, '03-landing-done.png'), fullPage: true });
});

test('Main screen IDLE on mount', async ({ page }) => {
  await page.goto('/main');
  await page.waitForTimeout(200);
  await page.screenshot({ path: path.join(SHOT_DIR, '04-main-idle.png'), fullPage: true });
});

test('Main screen auto-transitions to SUGGESTED', async ({ page }) => {
  await page.goto('/main');
  await page.waitForTimeout(1800);
  await page.screenshot({ path: path.join(SHOT_DIR, '05-main-suggested.png'), fullPage: true });
  await expect(page.getByText('そろそろだろ。')).toBeVisible();
});

test('SLOT animation frames', async ({ page }) => {
  await page.goto('/main');
  await page.waitForTimeout(1800);
  await page.getByRole('button', { name: /押す|めんどくさい/ }).click();

  for (let i = 0; i < 10; i++) {
    await page.waitForTimeout(250);
    await page.screenshot({
      path: path.join(SHOT_DIR, `06-slot-frame-${String(i).padStart(2, '0')}.png`),
    });
  }
});

test('DEAD state via debug', async ({ page }) => {
  await page.goto('/main');
  await page.locator('button:has-text("dead")').click();
  await page.waitForTimeout(1200);
  await page.screenshot({ path: path.join(SHOT_DIR, '07-dead.png'), fullPage: true });
  await expect(page.getByText('今月は、終わりだ。')).toBeVisible();
});

test('Complete screen', async ({ page }) => {
  await page.goto('/main');
  await page.waitForTimeout(1800);
  await page.getByRole('button', { name: /押す|めんどくさい/ }).click();
  await page.waitForURL(/\/main\/complete/, { timeout: 10_000 });
  await page.waitForTimeout(800);
  await page.screenshot({ path: path.join(SHOT_DIR, '08-complete.png'), fullPage: true });
  await expect(page.getByText('いい判断だ。')).toBeVisible();
});
