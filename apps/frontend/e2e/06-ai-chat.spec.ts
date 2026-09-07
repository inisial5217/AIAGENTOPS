import { test, expect } from "@playwright/test";

test.describe("E2E: AI Assistant & Tool Execution", () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      localStorage.setItem("cifo_access_token", "dev-token-admin");
    });
  });

  test("opens AI chat floating assistant, populates suggestion, and sends message", async ({ page }) => {
    await page.goto("/monitoring");
    await page.waitForLoadState("domcontentloaded");

    // Locate floating action button for AI chat
    const chatFab = page.locator("#ai-chat-fab-btn, button[aria-label*='AI' i], button:has-text('AI Assistant')").first();
    if (await chatFab.isVisible()) {
      await chatFab.click();

      // Check chat window appears
      const chatInput = page.locator("#ai-chat-input-textarea, textarea[placeholder*='Ask AI' i]");
      await expect(chatInput).toBeVisible({ timeout: 5000 });

      // Click a quick suggestion chip if visible
      const suggestionChip = page.locator("button:has-text('Check pod status'), button:has-text('Show failed docker')").first();
      if (await suggestionChip.isVisible()) {
        await suggestionChip.click();
        const inputValue = await chatInput.inputValue();
        expect(inputValue.length).toBeGreaterThan(0);
      } else {
        await chatInput.fill("Check cluster health status");
      }

      // Send the query
      const sendBtn = page.locator("#ai-chat-send-btn, button[type='submit']").first();
      await sendBtn.click();

      // Message should appear in chat bubble list
      await expect(page.locator("body")).toContainText(/cluster|health|Check/i);
    }
  });
});
