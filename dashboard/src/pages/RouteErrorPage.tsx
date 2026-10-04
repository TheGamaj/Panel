import { Box, Button, Heading, Text, VStack } from "@chakra-ui/react";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import {
	isRouteErrorResponse,
	useNavigate,
	useRouteError,
} from "react-router-dom";
import { recoverFromStaleChunk } from "../utils/chunkRecovery";
import PageFooter from "../components/PageFooter";

const routeErrorMessage = (error: unknown) => {
	if (isRouteErrorResponse(error)) {
		return error.statusText || `Request failed with status ${error.status}`;
	}
	if (error instanceof Error) {
		return error.message;
	}
	return "The page could not be loaded.";
};

/**
 * The screen shown when a route throws.
 *
 * It is the one Gamaj page that sits outside AppLayout, so it carries the same
 * surfaces, the same accent and the same closing author line as everything
 * else, and it is kept in its own module so it can be rendered on its own.
 */
export const RouteErrorPage = () => {
	const error = useRouteError();
	const navigate = useNavigate();
	const { t } = useTranslation();

	useEffect(() => {
		recoverFromStaleChunk(error);
	}, [error]);

	return (
		<Box
			alignItems="center"
			bg="panel.app"
			color="panel.text"
			display="flex"
			// Column, so the author line anchors to the bottom of the page just
			// as it does on every other screen.
			flexDirection="column"
			justifyContent="center"
			minH="100dvh"
			px={6}
			py={10}
			w="full"
		>
			<VStack align="start" spacing={4} maxW="720px" w="full">
				<Heading size="lg">{t("router.errorTitle")}</Heading>
				<Text color="panel.textSecondary">
					{t("router.errorDescription")}
				</Text>
				<Text
					bg="panel.surface"
					border="1px solid"
					borderColor="panel.border"
					borderRadius="8px"
					color="panel.text"
					fontFamily="mono"
					fontSize="sm"
					p={4}
					w="full"
					whiteSpace="pre-wrap"
				>
					{routeErrorMessage(error)}
				</Text>
				<Button
					bg="panel.accent"
					color="var(--gm-panel-bg)"
					_hover={{ opacity: 0.9 }}
					onClick={() => navigate("/")}
				>
					{t("router.backToDashboard")}
				</Button>
			</VStack>
			<PageFooter />
		</Box>
	);
};

export default RouteErrorPage;