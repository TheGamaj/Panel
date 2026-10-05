import { defineConfig, devices } from "@playwright/test";

/**
 * The brand sweep runs against the real dashboard, served by the real dev
 * server, with the mock API answering behind it. Nothing is stubbed: the
 * point is to read the colours the browser actually computes, which is the
 * one thing a source-level guard cannot do.
 *
 * GAMAJ_PW_CHANNEL picks a browser that is already installed instead of the
 * one Playwright downloads. Set it to `msedge` or `chrome` on a machine where
 * the download CDN is unreachable; CI leaves it unset and gets the pinned
 * browser, so the version under test is the same on every run.
 */
const PORT = Number(process.env.GAMAJ_PW_PORT ?? 4199);

export default defineConfig({
	testDir: "./tests",
	fullyParallel: false,
	workers: 1,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 1 : 0,
	reporter: [["list"]],
	timeout: 60_000,
	expect: { timeout: 15_000 },

	use: {
		baseURL: `http://127.0.0.1:${PORT}`,
		channel: process.env.GAMAJ_PW_CHANNEL as
			| "msedge"
			| "chrome"
			| undefined,
		viewport: { width: 1440, height: 900 },
		trace: "retain-on-failure",
	},

	projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],

	webServer: {
		command: `npx vite --host 127.0.0.1 --port ${PORT} --strictPort`,
		env: { GAMAJ_MOCK_API: "1" },
		url: `http://127.0.0.1:${PORT}/dashboard/users`,
		reuseExistingServer: !process.env.CI,
		timeout: 120_000,
	},
});
