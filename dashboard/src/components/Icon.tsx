import { Box, Text } from "@chakra-ui/react";
import type { FC, PropsWithChildren } from "react";

/**
 * The tones an icon mark may take.
 *
 * The old prop was a Chakra hue name, and the component interpolated it into
 * token paths (`${color}.400`) to stack three translucent copies of itself
 * into a coloured glow. That is decoration the identity does not allow, and it
 * is why the prop cannot be a bare colour name any more: a name like "blue"
 * says how it looks, not what it means.
 *
 * One mark, one flat fill, and the meaning comes from the tone.
 */
export type IconRole = "accent" | "danger" | "warning" | "success";

export type IconType = {
	tone?: IconRole;
	size?: number;
};

const MARK_BG: Record<IconRole, string> = {
	accent: "panel.accent",
	danger: "panel.danger",
	warning: "panel.warning",
	success: "panel.success",
};

const MARK_TEXT: Record<IconRole, string> = {
	accent: "panel.bg",
	danger: "panel.bg",
	warning: "panel.bg",
	success: "panel.bg",
};

export const Icon: FC<PropsWithChildren<IconType>> = ({
	children,
	tone = "accent",
	size = 36,
}) => {
	const baseSize = `${size}px`;
	return (
		<Box
			position="relative"
			width={baseSize}
			height={baseSize}
			display="flex"
			justifyContent="center"
			alignItems="center"
			flexShrink={0}
		>
			{/* The mark is a flat fill, not a glow: one surface, one border. */}
			<Box
				position="absolute"
				inset={0}
				bg={MARK_BG[tone]}
				borderRadius="4px"
				zIndex={1}
			/>
			<Text
				color={MARK_TEXT[tone]}
				position="relative"
				zIndex={2}
				lineHeight="1"
			>
				{children}
			</Text>
		</Box>
	);
};
