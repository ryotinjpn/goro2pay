import { test } from '@playwright/test';

test('Inspect slot strip transform', async ({ page }) => {
  await page.goto('/main');
  await page.waitForTimeout(200);
  await page.locator('button:has-text("めんどくさい"), button:has-text("押す。")').first().click();
  await page.waitForTimeout(300);

  for (let i = 0; i < 6; i++) {
    const t = await page.evaluate(() => {
      const all = Array.from(document.querySelectorAll('div')) as HTMLElement[];
      const el = all.find((d) =>
        Array.from(d.classList).some((c) => /strip/i.test(c)),
      );
      if (!el) {
        return {
          found: false,
          allClasses: Array.from(document.querySelectorAll('[class]'))
            .map((e) => (e as HTMLElement).className)
            .slice(0, 30),
        };
      }
      const cs = getComputedStyle(el);
      return {
        found: true,
        className: el.className,
        transform: cs.transform,
        animation: cs.animationName,
        animationDuration: cs.animationDuration,
        height: el.offsetHeight,
        text: el.innerText,
      };
    });
    console.log(`frame ${i}:`, JSON.stringify(t));
    await page.waitForTimeout(150);
  }
});
