import { ChakraProvider } from "@chakra-ui/react";
import * as React from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";
import { theme } from "../../chakra.config";
import { RouteErrorPage } from "./RouteErrorPage";

Object.assign(globalThis, { React });

vi.mock("react-i18next", () => ({
	useTranslation: () => ({
		t: (key: string) => key,
		i18n: { dir: () => "ltr", language: "en" },
	}),
}));

// The error screen reads the thrown value from the router; a stub is enough to
// render it, because the assertions are about the page around the message.
vi.mock("react-router-dom", () => ({
	useRouteError: () => new Error("boom"),
	useNavigate: () => () => undefined,
	isRouteErrorResponse: () => false,
}));

const render = () =>
	renderToStaticMarkup(
		React.createElement(
			ChakraProvider,
			{ theme },
			React.createElement(RouteErrorPage),
		),
	);

describe("RouteErrorPage", () => {
	it("closes with the author line like every other Gamaj page", () => {
		const html = render();
		expect(html).toContain("app.codedBy");
	});

	it("does not print the product version", () => {
		expect(render()).not.toContain("is.0.0.1");
	});

	it("still explains the failure and offers a way back", () => {
		const html = render();
		expect(html).toContain("router.errorTitle");
		expect(html).toContain("router.backToDashboard");
		expect(html).toContain("boom");
	});
});