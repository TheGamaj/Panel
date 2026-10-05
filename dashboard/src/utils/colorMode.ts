/**
 * The whole panel has exactly two colour modes: dark and light.
 *
 * There is no third mode and no accent picker. The identity is monochrome,
 * so anything that would let a hue or a custom palette reach the screen is
 * deliberately absent; the only thing a user chooses is which side of the
 * grey ramp they read against.
 *
 * Three pieces of state have to agree for a mode to take effect: the
 * Chakra colour mode, the `gm-theme-*` class and `data-theme` attribute the
 * stylesheet keys off, and the browser chrome colour. Applying a mode is
 * triplicated logic that must stay in step, so it lives here once.
 */
import { updateThemeColor } from "utils/themeColor";

export type ColorMode = "dark" | "light";

export const COLOR_MODE_STORAGE_KEY = "gm-theme";
const CHAKRA_COLOR_MODE_KEY = "chakra-ui-color-mode";

/** Anything that is not literally "light" is dark: the identity's default. */
export const normalizeColorMode = (value?: string | null): ColorMode =>
	value === "light" ? "light" : "dark";

export const getInitialColorMode = (): ColorMode => {
	try {
		return normalizeColorMode(
			localStorage.getItem(COLOR_MODE_STORAGE_KEY) ||
				localStorage.getItem(CHAKRA_COLOR_MODE_KEY),
		);
	} catch {
		return "dark";
	}
};

/**
 * Keys written by the theme picker that used to exist. They are cleared on
 * every mode change so a browser that still holds a saved accent or palette
 * drops it instead of carrying dead state around indefinitely.
 */
const RETIRED_THEME_KEYS = ["gm-accent", "gm-custom-themes"];

export const applyColorMode = (mode: ColorMode) => {
	const normalized = normalizeColorMode(mode);
	try {
		localStorage.setItem(COLOR_MODE_STORAGE_KEY, normalized);
		localStorage.setItem(CHAKRA_COLOR_MODE_KEY, normalized);
		for (const key of RETIRED_THEME_KEYS) localStorage.removeItem(key);
	} catch {}

	const targets = [document.documentElement, document.body].filter(
		Boolean,
	) as HTMLElement[];
	targets.forEach((target) => {
		target.classList.remove(
			"gm-theme-light",
			"gm-theme-dark",
			"chakra-ui-light",
			"chakra-ui-dark",
		);
		target.classList.add(`gm-theme-${normalized}`, `chakra-ui-${normalized}`);
		target.dataset.theme = normalized;
		target.style.colorScheme = normalized;
	});

	updateThemeColor(normalized);
	return normalized;
};