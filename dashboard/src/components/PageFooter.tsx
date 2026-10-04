import { Text, type HTMLChakraProps } from "@chakra-ui/react";
import { useTranslation } from "react-i18next";

/**
 * The author line that closes every Gamaj page.
 *
 * One component, used by the login screen and by every dashboard route, so
 * the line sits in exactly the same place and reads identically everywhere.
 * It is centered and sits at the very bottom of the page, directly under the
 * content, the way a colophon does.
 */
const PageFooter = (props: HTMLChakraProps<"footer">) => {
	const { t } = useTranslation();

	return (
		<Text
			as="footer"
			color="var(--gm-panel-text-muted)"
			fontSize="xs"
			letterSpacing="wide"
			textAlign="center"
			// Full width, so the text is centered on the page rather than on
			// the footer's own shrink-to-fit box.
			w="full"
			flexShrink={0}
			// mt="auto" anchors the line to the bottom of the page on short
			// pages; on long pages it simply follows the content.
			mt="auto"
			pt={8}
			pb={{ base: 4, md: 2 }}
			{...props}
		>
			{t("app.codedBy")}
		</Text>
	);
};

export default PageFooter;