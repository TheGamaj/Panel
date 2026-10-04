import { chakra, type HTMLChakraProps } from "@chakra-ui/react";
import type { FC } from "react";

/**
 * The Gamaj symbol, rendered inline.
 *
 * The mark is defined once in `assets/gamaj-mark.svg` and filled with
 * `currentColor`, so a single file serves both light and dark surfaces. That
 * only resolves when the SVG is part of the document rather than loaded
 * through an `<img>`, which is why the mark is drawn here. It also removes
 * the need for a CSS filter to invert the logo in dark mode.
 *
 * Geometry (from the Gamaj brand system): three equal squares in the
 * top-left, top-right and bottom-left cells of the unit grid, plus one smaller
 * square in the bottom-right cell for deliberate asymmetry.
 */
export const MARK_PATH =
	"M8 8h20v20H8V8zm28 0h20v20H36V8zM8 36h20v20H8V36zm28 8h12v12H36V44z";

export const GamajMark: FC<HTMLChakraProps<"svg">> = (props) => (
	<chakra.svg
		viewBox="0 0 64 64"
		fill="currentColor"
		role="img"
		aria-label="Gamaj"
		{...props}
	>
		<path d={MARK_PATH} />
	</chakra.svg>
);

export default GamajMark;
