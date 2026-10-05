/**
 * Mock API middleware for the dashboard dev server.
 *
 * The dashboard is a single-page app behind a router loader: without a
 * backend, `/auth/session` never resolves as active, the root loader
 * redirects to `/login`, and every authenticated route is unreachable. That
 * makes it impossible to check a change on those pages in a browser.
 *
 * Enabled only when GAMAJ_MOCK_API is set, and only on the dev server, so it
 * can never reach a production bundle:
 *
 *   GAMAJ_MOCK_API=1 npm run dev
 *
 * The point is to render real layouts against real-shaped data, not to
 * simulate the product. Anything the fixtures do not cover answers 404 so a
 * gap is visible rather than silently empty.
 */
import * as fixtures from "./mock-api.mjs";

const json = (res, status, body) => {
	const payload = JSON.stringify(body);
	res.statusCode = status;
	res.setHeader("Content-Type", "application/json; charset=utf-8");
	res.setHeader("Cache-Control", "no-store");
	res.end(payload);
};

const ok = (res, body) => json(res, 200, body);

/**
 * Which session the mock reports. Logged out is the state that makes the login
 * screen reachable at all, and it is also what an expired cookie looks like to
 * the app, so both are worth being able to open.
 */
const currentSession = () =>
	process.env.GAMAJ_MOCK_API_LOGGED_OUT === "1"
		? fixtures.loggedOutSession
		: fixtures.session;
const missing = (res, method, route) => {
	if (process.env.GAMAJ_MOCK_API_VERBOSE !== "1") {
		return json(res, 404, { error: `mock API has no fixture for ${method} ${route}` });
	}
	console.warn(`[mock-api] no fixture for ${method} ${route}`);
	return json(res, 404, { error: `mock API has no fixture for ${method} ${route}` });
};

/**
 * Match a request path against the fixture table. Exact paths win; a `prefix`
 * entry is only used when nothing else matched, so `/users/5` still reaches
 * the `/users` collection fixture rather than the single-user one.
 */
const ROUTES = [
	// Session and auth. This is the one that decides whether the app renders.
	{ method: "GET", path: "/auth/session", body: () => currentSession() },
	{ method: "POST", path: "/auth/login", body: () => fixtures.session },
	{ method: "POST", path: "/auth/logout", body: () => ({}) },
	{ method: "POST", path: "/auth/2fa/verify", body: () => fixtures.session },
	{ method: "POST", path: "/auth/2fa/setup", body: () => ({ secret: "MOCKSECRET", uri: "otpauth://mock" }) },
	{ method: "POST", path: "/auth/2fa/confirm", body: () => fixtures.session },
	{ method: "POST", path: "/auth/2fa", body: () => fixtures.session },
	{ method: "GET", path: "/auth/sessions", body: () => ({ sessions: fixtures.adminSessions }) },

	// Dashboard.
	//
	// There is deliberately no `/dashboard` entry: that is the name of the SPA
	// route itself, so a fixture for it would answer the page request with
	// JSON and the app would never boot. The dashboard page reads /system,
	// /system/metrics and /core instead.
	{ method: "GET", path: "/system", body: () => fixtures.system },
	{ method: "GET", path: "/system/metrics", body: () => fixtures.metrics },
	{ method: "GET", path: "/core", body: () => fixtures.core },
	{ method: "GET", path: "/core/config", body: () => fixtures.core },
	{ method: "POST", path: "/core/restart", body: () => fixtures.core },

	// Users.
	{ method: "GET", path: "/users", body: () => fixtures.users },
	{ method: "GET", path: "/users/onlines", body: () => fixtures.users.filter((u) => u.is_online) },
	{ method: "GET", prefix: "/users/", body: (_m, route) => {
		const username = decodeURIComponent(route.split("/").filter(Boolean)[1] || "");
		const found = fixtures.users.find((u) => u.username === username);
		return found || null;
	} },

	// Admins.
	{ method: "GET", path: "/admins", body: () => fixtures.admins },
	{ method: "GET", path: "/admin", body: () => fixtures.admins },
	{ method: "GET", path: "/admin/api-keys", body: () => fixtures.apiKeys },
	{ method: "GET", path: "/admin/permissions", body: () => ({ standard: [], sudo: [] }) },

	// Nodes.
	{ method: "GET", path: "/nodes", body: () => fixtures.nodes },
	{ method: "GET", path: "/nodes/usage", body: () => fixtures.metrics },
	{ method: "GET", path: "/nodes/metrics", body: () => fixtures.metrics },
	{ method: "GET", path: "/node", body: () => fixtures.nodes },

	// The service list is fetched at /v2/services and unwrapped from a
	// `{ services }` envelope, so the fixture must carry that shape or the
	// store sets `undefined` and every consumer throws on `.length`.
	{
		method: "GET",
		path: "/v2/services",
		body: () => ({ services: fixtures.services }),
	},
	// Hosts, services, haproxy.
	{ method: "GET", path: "/hosts", body: () => fixtures.hosts },
	{ method: "GET", path: "/services", body: () => fixtures.services },
	{ method: "GET", path: "/haproxy", body: () => fixtures.haproxy },
	{ method: "GET", path: "/haproxy/templates", body: () => fixtures.haproxy.templates },
	{ method: "GET", path: "/haproxy/preview", body: () => ({ html: "<pre>mock preview</pre>" }) },

	// Xray.
	{ method: "GET", path: "/inbounds", body: () => fixtures.inbounds },
	{ method: "GET", path: "/inbounds/full", body: () => fixtures.inbounds },
	{ method: "GET", path: "/outbounds", body: () => fixtures.outbounds },
	{ method: "GET", path: "/xray/config", body: () => fixtures.xray },
	{ method: "GET", path: "/xray-logs", body: () => fixtures.xrayLogs },

	// Settings.
	{ method: "GET", path: "/settings", body: () => fixtures.settingsAll },
	{ method: "GET", path: "/settings/all", body: () => fixtures.settingsAll },
	{ method: "GET", path: "/settings/panel", body: () => fixtures.settingsAll.panel },
	{ method: "GET", path: "/settings/placeholders", body: () => fixtures.placeholders },
	{ method: "GET", path: "/settings/subscriptions", body: () => fixtures.settingsAll.subscriptions },
	{ method: "GET", path: "/settings/external-apps", body: () => fixtures.externalApps },
	{ method: "GET", path: "/settings/telegram", body: () => fixtures.settingsAll.integrations },

	// External apps.
	{ method: "GET", path: "/external-apps", body: () => fixtures.externalApps },

	// Insights and audit.
	{ method: "GET", path: "/access-insights", body: () => fixtures.accessInsights },
	{ method: "GET", path: "/recent-actions", body: () => fixtures.recentActions },

	// Account.
	{ method: "GET", path: "/myaccount", body: () => fixtures.myAccount },

	// Misc reads that several pages touch.
	{ method: "GET", path: "/core/geo/templates", body: () => fixtures.geoTemplates },
	{ method: "GET", path: "/core/ips", body: () => ({ ips: [] }) },
	{ method: "GET", path: "/core/warp", body: () => fixtures.warp },
	{ method: "GET", path: "/core/warp/config", body: () => fixtures.warp.config },
	{ method: "GET", path: "/node-settings", body: () => fixtures.nodeSettings },
	{ method: "GET", path: "/tutorials", body: () => ({ tutorials: [] }) },
];

/** Strip the API base the dev client prefixes onto every request. */
const normalise = (url) => {
	const parsed = new URL(url || "/", "http://localhost");
	let route = parsed.pathname;
	if (route === "/api") route = "/";
	else if (route.startsWith("/api/")) route = route.slice(4);
	if (route.length > 1) route = route.replace(/\/+$/, "");
	return route;
};

/**
 * The panel API is mounted wherever VITE_BASE_API points. Left empty, which is
 * the default in .env.example, the client calls `/auth/session` at the origin
 * root; set to `/api/`, it calls `/api/auth/session`. Both have to resolve, so
 * a route is matched against the path with and without the `/api` prefix.
 */
const isApiRequest = (url) => {
	const { pathname } = new URL(url || "/", "http://localhost");
	return (
		pathname === "/api" ||
		pathname.startsWith("/api/") ||
		ROUTES.some(
			(entry) =>
				entry.path === normalise(url) ||
				(entry.prefix && normalise(url).startsWith(entry.prefix)),
		)
	);
};

export const mockApiMiddleware = (request, response, next) => {
	// Everything the fixtures do not name belongs to Vite: the HTML entry, the
	// module graph, the tutorial content, the favicon. Answering those here
	// would take the whole dev server down with a 404.
	if (!isApiRequest(request.url)) return next();

	const route = normalise(request.url);
	const method = (request.method || "GET").toUpperCase();
	const table = ROUTES.filter((entry) => entry.method === method);

	const exact = table.find((entry) => entry.path === route);
	const loose = exact ?? table.find((entry) => entry.prefix && route.startsWith(entry.prefix));
	if (!loose) return missing(response, method, route);

	const body = loose.body(method, route);
	if (body === null || body === undefined) return missing(response, method, route);
	return ok(response, body);
};

/** Report anything the fixtures do not cover, once per route. */
export const makeReporter = () => {
	const seen = new Set();
	return (method, route) => {
		const key = `${method} ${route}`;
		if (seen.has(key)) return;
		seen.add(key);
		process.stdout.write(`[mock-api] unmapped: ${key}\n`);
	};
};
