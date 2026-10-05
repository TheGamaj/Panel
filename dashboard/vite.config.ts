import react from "@vitejs/plugin-react";
import { visualizer } from "rollup-plugin-visualizer";
import {
	defineConfig,
	loadEnv,
	type Plugin,
	splitVendorChunkPlugin,
} from "vite";
import svgr from "vite-plugin-svgr";
import tsconfigPaths from "vite-tsconfig-paths";

// Loaded dynamically: the mock must never end up in a production bundle.
const loadMockApi = async () => {
	const module = await import("./tools/vite-plugin-mock-api.mjs");
	return module.mockApiMiddleware;
};

const tutorialDirectoryIndex = {
	name: "tutorial-directory-index",
	configureServer(server) {
		server.middlewares.use((request, _response, next) => {
			const url = new URL(request.url || "/", "http://localhost");
			if (
				url.pathname.startsWith("/tutorial-content/") &&
				url.pathname.endsWith("/")
			) {
				request.url = `${url.pathname}index.html${url.search}`;
			}
			next();
		});
	},
} satisfies Plugin;

/**
 * Serves fixture data for the panel API so authenticated routes can be opened
 * in a browser without a backend. Dev-server only, and only when
 * GAMAJ_MOCK_API is set, so it cannot affect a build.
 */
const mockApi = (enabled: boolean): Plugin => ({
	name: "gamaj-mock-api",
	apply: "serve",
	configureServer(server) {
		if (!enabled) return;
		let middleware;
		let pending;
		// The fixture module is ESM and the config is loaded before the server
		// exists, so it is imported here rather than at module scope.
		pending = loadMockApi().then((handler) => {
			middleware = handler;
		});
		server.middlewares.use((request, response, next) => {
			if (!middleware) {
				// A request that arrives before the import resolves is held
				// rather than passed through to the real (absent) API.
				pending.then(() => middleware(request, response, next));
				return;
			}
			middleware(request, response, next);
		});
	},
});

const getApiProxyConfig = (baseAPI?: string) => {
	if (!baseAPI || !/^https?:\/\//i.test(baseAPI)) {
		return undefined;
	}

	try {
		const parsed = new URL(baseAPI);
		const proxyPath =
			parsed.pathname && parsed.pathname !== "/"
				? parsed.pathname.replace(/\/$/, "")
				: "/api";
		const target = `${parsed.protocol}//${parsed.host}`;
		const rewrite =
			parsed.pathname && parsed.pathname !== "/"
				? undefined
				: (path: string) => path.replace(/^\/api(?=\/|$)/, "");

		return {
			proxyPath,
			options: {
				target,
				changeOrigin: true,
				secure: true,
				rewrite,
			},
		};
	} catch {
		return undefined;
	}
};

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd(), "");
	const apiProxy = getApiProxyConfig(env.VITE_BASE_API);

	return {
		plugins: [
			tutorialDirectoryIndex,
			tsconfigPaths(),
			react({
				include: "**/*.tsx",
			}),
			svgr(),
			...(env.ANALYZE === "true" ? [visualizer()] : []),
			splitVendorChunkPlugin(),
			mockApi(env.GAMAJ_MOCK_API === "1" || env.GAMAJ_MOCK_API === "true"),
		],
		server: apiProxy
			? {
					proxy: {
						[apiProxy.proxyPath]: apiProxy.options,
					},
				}
			: undefined,
		build: {
			outDir: "build",
			assetsDir: "statics",
			rollupOptions: {
				onwarn(warning, warn) {
					if (
						typeof warning.message === "string" &&
						warning.message.includes(
							"Module level directives cause errors when bundled",
						)
					) {
						return;
					}
					warn(warning);
				},
			},
		},
	};
});
