/**
 * Keeps the browser chrome colour in step with the active colour mode.
 *
 * The Gamaj identity is monochrome, so the theme colour is the pure black or
 * pure white surface rather than a tinted approximation.
 */
export const updateThemeColor = (themeName: string, fallback?: string) => {
	const el = document.querySelector('meta[name="theme-color"]');
	const map: Record<string, string> = {
		dark: "#000000",
		light: "#FFFFFF",
		custom: fallback || "#000000",
	};
	const color = fallback || map[themeName] || map.dark;
	el?.setAttribute("content", color);
};
