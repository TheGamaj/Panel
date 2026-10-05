import { MenuItem, useColorMode } from "@chakra-ui/react";
import { CheckIcon, MoonIcon, SunIcon } from "@heroicons/react/24/outline";
import type { FC, ReactElement } from "react";
import { useTranslation } from "react-i18next";
import {
	applyColorMode,
	normalizeColorMode,
	type ColorMode,
} from "utils/colorMode";

const Moon = () => <MoonIcon width={16} height={16} />;
const Sun = () => <SunIcon width={16} height={16} />;

/**
 * The panel's only appearance setting: dark or light.
 *
 * This used to be a ThemeSelector offering accent colours, built-in
 * presets and importable custom palettes. All of that is gone. The
 * identity is monochrome, so a hue picker could only produce a panel that
 * looked like a different product, and the picker itself carried a large
 * amount of state — a modal, an export/import format, per-mode background,
 * surface and primary values — for a choice the brand does not have.
 */
export const ColorModeMenuItems: FC = () => {
	const { t } = useTranslation();
	const { colorMode, setColorMode } = useColorMode();
	const activeMode = normalizeColorMode(colorMode);

	const options: Array<{
		key: ColorMode;
		label: string;
		icon: ReactElement;
	}> = [
		{ key: "dark", label: t("theme.dark"), icon: <Moon /> },
		{ key: "light", label: t("theme.light"), icon: <Sun /> },
	];

	const select = (mode: ColorMode) => {
		applyColorMode(mode);
		setColorMode(mode);
	};

	return (
		<>
			{options.map((option) => {
				const isActive = option.key === activeMode;
				return (
					<MenuItem
						key={option.key}
						className="gm-color-mode-menu-item"
						icon={option.icon}
						onClick={() => select(option.key)}
						aria-current={isActive ? "true" : undefined}
						bg="transparent"
						_hover={{ bg: "panel.elevated" }}
						_active={{ bg: "transparent" }}
						_focus={{ bg: "transparent" }}
						_focusVisible={{ bg: "panel.elevated" }}
					>
						<span>{option.label}</span>
						{isActive && (
							<CheckIcon width={16} color="var(--gm-panel-accent)" />
						)}
					</MenuItem>
				);
			})}
		</>
	);
};

export default ColorModeMenuItems;