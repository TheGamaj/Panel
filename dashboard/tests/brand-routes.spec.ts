import { expect, test, type Page } from "@playwright/test";

/**
 * The brand invariant, checked the only way it can honestly be checked: by
 * reading the colours the browser resolves.
 *
 * The source-level guard cannot see a `color-mix()` between two palette
 * values, and it cannot see a value inherited from a custom property that
 * resolves to something between two ramp steps. That is how the Chakra
 * neutral scale came to default every border in light mode to #CECECE, and
 * how the bulk action bar resolved to #F0F0F0 — both invisible to a grep.
 * Computed styles catch the whole class.
 *
 * Two things are asserted per route, in both colour modes:
 *   1. the author line is present, centred, and the last thing on the page;
 *   2. no element resolves a colour from outside the Gamaj ramp.
 *
 * Console noise is deliberately not asserted. The app already logs React
 * DOM-nesting warnings and the mock API does not serve the tutorial meta
 * endpoint, so a "no console errors" rule here would be permanently red for
 * reasons that have nothing to do with the brand. Those are tracked on their
 * own.
 */

const ROUTES = [
	"",
	"users",
	"bulk-actions",
	"admins",
	"usage",
	"myaccount",
	"settings",
	"node-settings",
	"services",
	"hosts",
	"haproxy",
	"xray-settings",
	"xray-logs",
	"access-insights",
	"recent-actions",
	"phpmyadmin",
	"external-apps",
	"placeholders",
	"api-docs",
	"tutorials",
] as const;

const COLOUR_MODES = ["dark", "light"] as const;

/** The Gamaj ramp, plus the three status hues that carry meaning. */
const PALETTE = new Set([
	"#0A0A0A",
	"#111111",
	"#1A1A1A",
	"#333333",
	"#4A4A4A",
	"#666666",
	"#999999",
	"#CCCCCC",
	"#E5E5E5",
	"#F5F5F5",
	"#FAFAFA",
	"#FFFFFF",
	"#000000",
	"#C62828",
	"#8D6E00",
	"#1B5E20",
]);

/** Sub-pixel rounding at the edges of the layout is not a layout fault. */
const CENTRE_TOLERANCE_PX = 1;

type ColourHit = {
	element: string;
	property: string;
	resolved: string;
};

/**
 * Resolve one computed colour to an uppercase hex, or null when it is not a
 * solid colour worth judging. Alpha is skipped on purpose: a wash at 8%
 * white is a legal way to lighten a surface and is not the identity failing.
 */
function toHex(value: string): string | null {
	const colour = value.trim();
	if (colour === "transparent" || colour === "") return null;

	const rgb = colour.match(/^rgba?\(([^)]+)\)$/i);
	if (rgb) {
		const parts = rgb[1]
			.split(/[\s,/]+/)
			.filter(Boolean)
			.map(Number);
		if (parts.length > 3 && (parts[3] === 0 || parts[3] < 1)) return null;
		return toHexFrom255(parts[0], parts[1], parts[2]);
	}

	// Chrome serialises a resolved color-mix() as color(srgb r g b / a).
	const srgb = colour.match(/^color\(srgb\s+([^)]+)\)$/i);
	if (srgb) {
		const parts = srgb[1]
			.split(/[\s,/]+/)
			.filter(Boolean)
			.map(Number);
		if (parts.length > 3 && (parts[3] === 0 || parts[3] < 1)) return null;
		return toHexFromUnit(parts[0], parts[1], parts[2]);
	}

	if (/^#[0-9A-F]{6}$/i.test(colour)) return colour.toUpperCase();
	return colour.toUpperCase();
}

const toHexFrom255 = (r: number, g: number, b: number) =>
	`#${[r, g, b]
		.map((v) => Math.round(v).toString(16).padStart(2, "0"))
		.join("")}`.toUpperCase();

const toHexFromUnit = (r: number, g: number, b: number) =>
	toHexFrom255(r * 255, g * 255, b * 255);

const COLOUR_PROPERTIES = [
	"color",
	"backgroundColor",
	"borderTopColor",
	"borderBottomColor",
	"borderLeftColor",
	"borderRightColor",
] as const;

/** The browser's own extension UI is not part of the application. */
function isForeignOverlay(element: HTMLElement): boolean {
	const style = window.getComputedStyle(element);
	return style.zIndex === "2147483647" && element.textContent?.trim() === "Freebuff";
}

async function collectOffPaletteColours(page: Page): Promise<ColourHit[]> {
	return page.evaluate(
		({ palette, properties, isForeignOverlaySource }) => {
			// The predicate cannot cross the page boundary, so it is rebuilt
			// here from the same rule the caller uses.
			const isForeign = new Function(
				"element",
				`return (${isForeignOverlaySource})(element);`,
			) as (element: HTMLElement) => boolean;

			const toHex = (value: string): string | null => {
				const colour = value.trim();
				if (colour === "transparent" || colour === "") return null;
				const rgb = colour.match(/^rgba?\(([^)]+)\)$/i);
				if (rgb) {
					const parts = rgb[1]
						.split(/[\s,/]+/)
						.filter(Boolean)
						.map(Number);
					if (parts.length > 3 && (parts[3] === 0 || parts[3] < 1)) {
						return null;
					}
					return toHexFrom255(parts[0], parts[1], parts[2]);
				}
				const srgb = colour.match(/^color\(srgb\s+([^)]+)\)$/i);
				if (srgb) {
					const parts = srgb[1]
						.split(/[\s,/]+/)
						.filter(Boolean)
						.map(Number);
					if (parts.length > 3 && (parts[3] === 0 || parts[3] < 1)) {
						return null;
					}
					return toHexFrom255(parts[0] * 255, parts[1] * 255, parts[2] * 255);
				}
				return colour.toUpperCase();
			};

			const toHexFrom255 = (r: number, g: number, b: number) =>
				`#${[r, g, b]
					.map((v) => Math.round(v).toString(16).padStart(2, "0"))
					.join("")}`.toUpperCase();

			const allowed = new Set(palette as string[]);
			const hits: Array<{
				element: string;
				property: string;
				resolved: string;
			}> = [];

			for (const element of Array.from(document.querySelectorAll("*"))) {
				if (isForeign(element as HTMLElement)) continue;
				const style = window.getComputedStyle(element);
				for (const property of properties as string[]) {
					const resolved = toHex(style[property as never]);
					if (resolved === null) continue;
					if (allowed.has(resolved)) continue;
					const className =
						typeof element.className === "string" &&
						element.className.length > 0
							? `.${element.className.trim().split(/\s+/).join(".")}`
							: "";
					hits.push({
						element: `${element.tagName.toLowerCase()}${className}`.slice(0, 60),
						property,
						resolved,
					});
				}
			}
			return hits;
		},
		{
			palette: [...PALETTE],
			properties: [...COLOUR_PROPERTIES],
			isForeignOverlaySource: isForeignOverlay.toString(),
		},
	);
}

async function readFooter(page: Page) {
	return page.evaluate(() => {
		const footer = document.querySelector("footer");
		if (!footer) return null;
		const parent = footer.parentElement as HTMLElement;
		const style = window.getComputedStyle(footer);
		const rect = footer.getBoundingClientRect();
		const parentRect = parent.getBoundingClientRect();
		return {
			text: footer.textContent?.trim() ?? "",
			textAlign: style.textAlign,
			isLastInParent: parent.lastElementChild === footer,
			// The parent's client box, not its border box: main scrolls, and
			// a scrollbar is not a layout offset. Measuring against the
			// border box reports a phantom half-scrollbar of misalignment.
			offsetFromCentre: (rect.left + rect.right) / 2 - (parentRect.left + parent.clientWidth / 2),
		};
	});
}

async function isErrorPage(page: Page): Promise<boolean> {
	return page.evaluate(
		() =>
			/something went wrong|unexpected application error|page not found/i.test(
				document.body.innerText,
			),
	);
}

for (const mode of COLOUR_MODES) {
	test.describe(`colour mode: ${mode}`, () => {
		for (const route of ROUTES) {
			const path = route === "" ? "/dashboard" : `/dashboard/${route}`;

			test(`${path} keeps the author line and stays on the palette`, async ({
				page,
			}) => {
				await page.goto(path, { waitUntil: "domcontentloaded" });

				// The mode is applied before the first paint, so setting it and
				// reloading keeps the assertion honest: it tests the shipped
				// boot path, not a runtime poke at the DOM.
				await page.evaluate((wanted) => {
					localStorage.setItem("gm-theme", wanted);
					localStorage.setItem("chakra-ui-color-mode", wanted);
				}, mode);
				await page.reload({ waitUntil: "domcontentloaded" });
				await page.waitForLoadState("networkidle");

				expect(
					await page.evaluate(
						() => document.documentElement.getAttribute("data-theme"),
					),
				).toBe(mode);

				expect(await isErrorPage(page), `${path} rendered an error page`).toBe(
					false,
				);

				const footer = await readFooter(page);
				expect(footer, `${path} has no footer at all`).not.toBeNull();
				expect(footer?.text).toBe("Coded by AsliCode");
				expect(footer?.textAlign).toBe("center");
				expect(footer?.isLastInParent, `${path} footer is not the last element`).toBe(
					true,
				);
				expect(
					Math.abs(footer?.offsetFromCentre ?? Infinity),
					`${path} footer is off centre by ${footer?.offsetFromCentre}px`,
				).toBeLessThanOrEqual(CENTRE_TOLERANCE_PX);

				const offPalette = await collectOffPaletteColours(page);
				expect(
					offPalette,
					`${path} resolved ${offPalette.length} colour(s) outside the Gamaj ramp: ` +
						offPalette
							.slice(0, 6)
							.map((h) => `${h.element} ${h.property} -> ${h.resolved}`)
							.join("; "),
				).toEqual([]);
			});
		}
	});
}
