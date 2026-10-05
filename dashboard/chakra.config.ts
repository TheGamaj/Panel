import { extendTheme } from "@chakra-ui/react";
import { mode, type StyleFunctionProps } from "@chakra-ui/theme-tools";

/**
 * Gamaj panel theme.
 *
 * Every colour, radius, font size and transition below comes from the Gamaj
 * brand system: a strictly monochrome palette (black, white, and the neutral
 * grey ramp), the 4px spacing scale, the three corner radii (2 / 4 / 8), and a
 * single calm transition curve. There are no gradients, no glows and no
 * shadows on resting surfaces.
 *
 * Dark is the default mode. Light mode is the same system with the ramp
 * inverted, so the two read as one identity rather than two themes.
 */
/**
 * Build a 50-950 scale out of one brand colour.
 *
 * Chakra recipes reach for `${hue}.500` and its neighbours, so a hue has to
 * answer every step. The steps are derived by mixing the hue with white in
 * the light half of the scale and with the panel background in the dark half,
 * which keeps the same hue at every step instead of inventing tints.
 */
const gamajHue = (hue: string) => {
	const scale: Record<number, string> = {};
	const steps: Array<[number, number]> = [
		[50, 92],
		[100, 84],
		[200, 68],
		[300, 52],
		[400, 34],
		[500, 0],
		[600, 12],
		[700, 26],
		[800, 42],
		[900, 58],
		[950, 70],
	];
	for (const [step, mix] of steps) {
		// `mix` is the percentage of white (below 500) or of the near-black
		// canvas (above it) blended into the hue.
		const towards =
			mix === 0
				? null
				: step < 500
					? "255, 255, 255"
					: "10, 10, 10";
		const amount = mix === 0 ? 0 : mix / 100;
		scale[step] = towards
			? `color-mix(in srgb, ${hue} ${Math.round((1 - amount) * 100)}%, rgb(${towards}))`
			: hue;
	}
	return scale;
};

const sharedThemeConfig = {
	config: {
		initialColorMode: "dark",
		useSystemColorMode: false,
	},
	direction: "ltr" as const,

	// The identity forbids decorative shadow; focus is shown with a ring.
	shadows: { outline: "0 0 0 2px var(--gamaj-gray-500)" },

	fonts: {
		body: `Arad,Inter,"SF Pro Display",-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Oxygen,Ubuntu,Cantarell,"Fira Sans","Droid Sans","Helvetica Neue","Apple Color Emoji","Segoe UI Emoji","Segoe UI Symbol",sans-serif`,
		heading: `Inter,"SF Pro Display",Arad,-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,Arial,sans-serif`,
		mono: `"SF Mono","Fira Code",Consolas,ui-monospace,monospace`,
	},

	// The type scale from the identity: display, H1-H4, body, caption.
	fontSizes: {
		xs: "11px",
		sm: "13px",
		base: "15px",
		md: "15px",
		lg: "18px",
		xl: "24px",
		"2xl": "28px",
		"3xl": "32px",
		"4xl": "48px",
	},

	fontWeights: {
		normal: 400,
		medium: 500,
		semibold: 600,
		bold: 600,
	},

	// Three radii, and nothing in between.
	radii: {
		none: "0",
		sm: "var(--gm-radius-sm)",
		md: "var(--gm-radius-md)",
		lg: "var(--gm-radius-lg)",
		xl: "var(--gm-radius-lg)",
		"2xl": "var(--gm-radius-lg)",
		"3xl": "var(--gm-radius-lg)",
		full: "var(--gm-radius-lg)",
	},

	space: {
		px: "1px",
		0.5: "2px",
		1: "4px",
		1.5: "6px",
		2: "8px",
		2.5: "10px",
		3: "12px",
		3.5: "14px",
		4: "16px",
		5: "24px",
		6: "32px",
		7: "48px",
		8: "64px",
		9: "96px",
		10: "128px",
	},

	sizes: {
		container: {
			sm: "640px",
			md: "768px",
			lg: "1024px",
			xl: "1120px",
		},
	},

	colors: {
		"light-border": "var(--gamaj-gray-200)",
		// Semantic panel surface roles, resolved through the CSS variables so
		// a colour-mode switch needs no component changes.
		panel: {
			app: "var(--gm-panel-bg)",
			main: "var(--gm-panel-main)",
			sidebar: "var(--gm-panel-sidebar)",
			surface: "var(--gm-panel-surface)",
			elevated: "var(--gm-panel-elevated)",
			border: "var(--gm-panel-border)",
			borderStrong: "var(--gm-panel-border-strong)",
			text: "var(--gm-panel-text)",
			textSecondary: "var(--gm-panel-text-secondary)",
			textMuted: "var(--gm-panel-text-muted)",
			accent: "var(--gm-panel-accent)",
			accentHover: "var(--gm-panel-accent-hover)",
			rowHover: "var(--gm-panel-row-hover)",
			rowSelected: "var(--gm-panel-row-selected)",
			inset: "var(--gm-panel-inset)",
			scrim: "var(--gm-panel-scrim)",
			danger: "var(--gm-danger)",
			dangerSubtle: "var(--gm-danger-subtle)",
			warning: "var(--gm-warning)",
			success: "var(--gm-success)",
		},
		// The brand neutral ramp, exposed to Chakra so component defaults can
		// reference the exact identity values.
		gamaj: {
			black: "var(--gamaj-black)",
			white: "var(--gamaj-white)",
			50: "var(--gamaj-gray-50)",
			100: "var(--gamaj-gray-100)",
			200: "var(--gamaj-gray-200)",
			300: "var(--gamaj-gray-300)",
			400: "var(--gamaj-gray-400)",
			500: "var(--gamaj-gray-500)",
			600: "var(--gamaj-gray-600)",
			700: "var(--gamaj-gray-700)",
			800: "var(--gamaj-gray-800)",
			900: "var(--gamaj-gray-900)",
			950: "var(--gamaj-gray-950)",
		},
		// Chakra's chromatic scales are redefined here rather than left alone.
		//
		// Two hundred call sites pass colorScheme="red" / "green" / "blue" to
		// Badge, Tag, Alert and Button, and those recipes resolve the name
		// against this palette. Pointing the scales at the Gamaj roles means
		// every one of those sites lands on the identity without touching the
		// call sites, and a chromatic hex can no longer reach the screen by
		// accident. The status hues keep their meaning; the purely decorative
		// ones collapse onto the accent and the neutral ramp.
		red: gamajHue("--gm-danger"),
		orange: gamajHue("--gm-warning"),
		yellow: gamajHue("--gm-warning"),
		green: gamajHue("--gm-success"),
		blue: gamajHue("--gm-panel-accent"),
		cyan: gamajHue("--gm-panel-accent"),
		teal: gamajHue("--gm-panel-accent"),
		purple: gamajHue("--gm-panel-accent"),
		pink: gamajHue("--gm-panel-accent"),
		gray: gamajHue("var(--gamaj-gray-500)"),
		bg: {
			light: "var(--bg-light)",
			dark: "var(--bg-dark)",
		},
		surface: {
			light: "var(--surface-light)",
			dark: "var(--surface-dark)",
		},
		// The accent scale is the neutral ramp: primary actions are the black
		// end in light mode and the white end in dark mode.
		primary: {
			50: "var(--primary-50)",
			100: "var(--primary-100)",
			200: "var(--primary-200)",
			300: "var(--primary-300)",
			400: "var(--primary-400)",
			500: "var(--primary-500)",
			600: "var(--primary-600)",
			700: "var(--primary-700)",
			800: "var(--primary-800)",
			900: "var(--primary-900)",
		},
	},

	styles: {
		global: {
			// Flat overlay: the identity uses no blur or tint behind dialogs.
			".chakra-modal__overlay": {
				bg: "panel.scrim !important",
				backdropFilter: "none !important",
				WebkitBackdropFilter: "none !important",
			},
			".chakra-modal__content": {
				backgroundColor: "var(--gm-panel-surface) !important",
				color: "var(--gm-panel-text) !important",
				borderColor: "var(--gm-panel-border) !important",
				borderRadius: "var(--gm-radius-lg) !important",
				boxShadow: "none !important",
			},
			// A single place for the mode-scoped surface variables, so Chakra
			// components and plain CSS always read the same values.
			":root, .gm-theme-dark, .chakra-ui-dark": {
				"--bg-light": "var(--gamaj-gray-950)",
				"--bg-dark": "var(--gamaj-gray-950)",
				"--surface-light": "var(--gamaj-gray-900)",
				"--surface-dark": "var(--gamaj-gray-900)",
			},
			".gm-theme-light, .chakra-ui-light": {
				"--bg-light": "var(--gamaj-white)",
				"--bg-dark": "var(--gamaj-white)",
				"--surface-light": "var(--gamaj-white)",
				"--surface-dark": "var(--gamaj-white)",
			},
			body: {
				backgroundColor: "panel.main",
				color: "panel.text",
				fontSize: "base",
			},
			"[data-theme='dark'] body, .chakra-ui-dark body": {
				backgroundColor: "panel.main",
				color: "panel.text",
			},
			"[data-theme='light'] body, .chakra-ui-light body": {
				backgroundColor: "panel.main",
				color: "panel.text",
			},
			// The wordmark, section labels and buttons all share the identity's
			// tracked uppercase treatment.
			".gm-wordmark": {
				fontFamily: "heading",
				fontWeight: "semibold",
				letterSpacing: "0.14em",
				textTransform: "uppercase",
			},
			".gm-eyebrow": {
				fontSize: "11px",
				fontWeight: "medium",
				letterSpacing: "0.12em",
				textTransform: "uppercase",
				color: "panel.textMuted",
			},
		},
	},

	components: {
		Card: {
			baseStyle: {
				container: {
					bg: "panel.surface",
					borderWidth: "1px",
					borderColor: "panel.border",
					boxShadow: "none",
					borderRadius: "var(--gm-radius-md)",
				},
			},
		},
		Badge: {
			baseStyle: {
				borderRadius: "var(--gm-radius-sm)",
				fontSize: "11px",
				fontWeight: "medium",
				letterSpacing: "0.06em",
				textTransform: "uppercase",
				px: "2px",
				py: "1px",
			},
		},
		Tag: {
			baseStyle: {
				container: {
					borderRadius: "var(--gm-radius-sm)",
					fontSize: "11px",
					fontWeight: "medium",
					letterSpacing: "0.06em",
				},
			},
		},
		Modal: {
			baseStyle: (props: StyleFunctionProps) => ({
				dialog: {
					bg: mode("panel.surface", "panel.surface")(props),
					borderWidth: "1px",
					borderColor: mode("panel.border", "panel.border")(props),
					borderRadius: "var(--gm-radius-lg)",
					boxShadow: "none",
				},
				header: {
					fontSize: "lg",
					fontWeight: "semibold",
					letterSpacing: "-0.015em",
					borderBottomWidth: "1px",
					borderColor: mode("panel.border", "panel.border")(props),
				},
				body: {
					fontSize: "base",
				},
				footer: {
					borderTopWidth: "1px",
					borderColor: mode("panel.border", "panel.border")(props),
				},
			}),
		},
		Drawer: {
			baseStyle: (props: StyleFunctionProps) => ({
				dialog: {
					bg: mode("panel.surface", "panel.surface")(props),
					borderColor: mode("panel.border", "panel.border")(props),
					borderWidth: "0",
				},
			}),
		},
		Menu: {
			baseStyle: (props: StyleFunctionProps) => {
				const hoverBg = mode("panel.elevated", "panel.elevated")(props);
				return {
					list: {
						bg: mode("panel.surface", "panel.surface")(props),
						borderWidth: "1px",
						borderColor: mode("panel.border", "panel.border")(props),
						boxShadow: "none",
						borderRadius: "var(--gm-radius-md)",
					},
					item: {
						bg: "transparent !important",
						color: mode("panel.text", "panel.text")(props),
						_fontSize: "sm",
						_hover: { bg: `${hoverBg} !important` },
						_focus: { bg: `${hoverBg} !important` },
						_active: { bg: `${hoverBg} !important` },
					},
				};
			},
		},
		Popover: {
			baseStyle: (props: StyleFunctionProps) => ({
				content: {
					bg: mode("panel.surface", "panel.surface")(props),
					borderWidth: "1px",
					borderColor: mode("panel.border", "panel.border")(props),
					boxShadow: "none",
					borderRadius: "var(--gm-radius-md)",
				},
				header: {
					borderBottomWidth: "1px",
					borderColor: mode("panel.border", "panel.border")(props),
				},
				footer: {
					borderTopWidth: "1px",
					borderColor: mode("panel.border", "panel.border")(props),
				},
			}),
		},
		Accordion: {
			baseStyle: (props: StyleFunctionProps) => ({
				container: {
					borderTopWidth: "0",
					borderBottomWidth: "1px",
					borderColor: mode("panel.border", "panel.border")(props),
					_last: { borderBottomWidth: "1px" },
				},
				button: {
					bg: "transparent",
					_fontSize: "base",
					_hover: { bg: mode("panel.elevated", "panel.elevated")(props) },
					_expanded: {
						bg: mode("panel.elevated", "panel.elevated")(props),
					},
				},
				panel: {
					bg: mode("panel.surface", "panel.surface")(props),
				},
			}),
		},
		Alert: {
			baseStyle: {
				container: {
					borderRadius: "var(--gm-radius-md)",
					borderWidth: "1px",
					fontSize: "sm",
				},
			},
		},
		Tooltip: {
			baseStyle: {
				bg: "gamaj.950",
				color: "gamaj.white",
				borderRadius: "var(--gm-radius-sm)",
				fontSize: "xs",
				px: "2",
				py: "1",
			},
		},
		Select: {
			baseStyle: {
				field: {
					bg: "panel.surface",
					color: "panel.text",
					borderRadius: "var(--gm-radius-md)",
					_dark: {
						borderColor: "panel.borderStrong",
					},
					_light: {
						borderColor: "panel.borderStrong",
					},
				},
			},
		},
		FormHelperText: {
			baseStyle: {
				fontSize: "xs",
				color: "panel.textMuted",
			},
		},
		FormLabel: {
			baseStyle: {
				fontSize: "sm",
				fontWeight: "medium",
				color: "panel.textSecondary",
				mb: "1",
				_dark: { color: "panel.textSecondary" },
				_light: { color: "panel.textSecondary" },
			},
		},
		Input: {
			baseStyle: {
				addon: {
					bg: "panel.elevated",
					borderColor: "panel.border",
					_dark: {
						borderColor: "panel.borderStrong",
						_placeholder: { color: "panel.textMuted" },
					},
					_light: {
						borderColor: "panel.borderStrong",
						_placeholder: { color: "panel.textMuted" },
					},
				},
				field: {
					bg: "panel.surface",
					color: "panel.text",
					borderRadius: "var(--gm-radius-md)",
					_focusVisible: {
						boxShadow: "none",
						borderColor: "gamaj.black",
						outlineColor: "gamaj.black",
					},
					_dark: {
						borderColor: "panel.borderStrong",
						_focusVisible: { borderColor: "gamaj.white" },
						_disabled: {
							color: "panel.textMuted",
							borderColor: "panel.border",
						},
						_placeholder: { color: "panel.textMuted" },
					},
					_light: {
						borderColor: "panel.borderStrong",
						_focusVisible: { borderColor: "gamaj.black" },
						_disabled: {
							color: "panel.textMuted",
							borderColor: "panel.border",
						},
						_placeholder: { color: "panel.textMuted" },
					},
				},
			},
		},
		// Tables follow the identity: hairline separators, an uppercase muted
		// header row, and no zebra fill or hover shadow.
		Table: {
			baseStyle: {
				table: {
					borderCollapse: "collapse",
					borderSpacing: 0,
					fontSize: "sm",
				},
				thead: {
					borderBottomColor: "panel.border",
				},
				th: {
					borderColor: "panel.border",
					borderBottomColor: "panel.border",
					borderTopWidth: "1px",
					color: "panel.textMuted",
					fontSize: "xs",
					fontWeight: "medium",
					letterSpacing: "0.06em",
					textTransform: "uppercase",
					px: "3",
					py: "2.5",
					_first: { borderLeftWidth: "1px" },
					_last: { borderRightWidth: "1px" },
				},
				td: {
					transition: "background var(--gm-transition)",
					borderColor: "panel.border",
					borderBottomColor: "panel.border",
					px: "3",
					py: "2.5",
					_first: { borderLeftWidth: "1px" },
					_last: { borderRightWidth: "1px" },
					_dark: { borderColor: "panel.border" },
					_light: { borderColor: "panel.border" },
				},
				tr: {
					"&.interactive": {
						cursor: "pointer",
						_hover: {
							"& > td": { bg: "panel.rowHover" },
						},
					},
					_last: { "& > td": { borderBottomWidth: "1px" } },
				},
			},
		},
		Button: {
			baseStyle: {
				_fontSize: "sm",
				_fontWeight: "medium",
				borderRadius: "var(--gm-radius-md)",
				transition: "background var(--gm-transition), border-color var(--gm-transition), color var(--gm-transition)",
			},
			variants: {
				solid: (props: StyleFunctionProps) => ({
					bg: mode("gamaj.black", "gamaj.white")(props),
					color: mode("gamaj.white", "gamaj.black")(props),
					_hover: {
						bg: mode("gamaj.800", "gamaj.200")(props),
						_disabled: {
							bg: mode("gamaj.800", "gamaj.200")(props),
						},
					},
					_active: { bg: mode("gamaj.900", "gamaj.100")(props) },
				}),
				outline: (props: StyleFunctionProps) => ({
					borderColor: mode("gamaj.300", "gamaj.700")(props),
					color: mode("gamaj.black", "gamaj.white")(props),
					_hover: {
						bg: mode("gamaj.50", "gamaj.900")(props),
						borderColor: mode("gamaj.black", "gamaj.white")(props),
					},
					_active: {
						bg: mode("gamaj.100", "gamaj.800")(props),
					},
				}),
				ghost: (props: StyleFunctionProps) => ({
					color: mode("gamaj.500", "gamaj.400")(props),
					_hover: {
						bg: mode("gamaj.100", "gamaj.800")(props),
						color: mode("gamaj.black", "gamaj.white")(props),
					},
					_active: {
						bg: mode("gamaj.200", "gamaj.800")(props),
					},
				}),
			},
			sizes: {
				sm: { h: "28px", px: "10px", fontSize: "xs" },
				md: { h: "36px", px: "14px" },
				lg: { h: "44px", px: "20px" },
			},
			defaultProps: {
				size: "md",
				colorScheme: "gray",
			},
		},
	},
};

export const theme = extendTheme(sharedThemeConfig);
export const rtlTheme = extendTheme({ ...sharedThemeConfig, direction: "rtl" });
