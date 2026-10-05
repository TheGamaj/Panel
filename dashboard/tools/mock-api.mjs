/**
 * Fixture data for the mock API.
 *
 * Kept apart from the middleware so the shapes are easy to read and to extend.
 * Everything here is fake but structurally faithful: the dashboard does real
 * work against these responses (sorting, summing, rendering charts), so empty
 * arrays would hide layout bugs rather than expose them.
 */

export const now = "2026-01-15T12:00:00Z";

export const admin = {
	id: 1,
	username: "asli",
	password: "",
	is_sudo: true,
	role: "full_access",
	created_at: "2025-06-01T09:00:00Z",
	used_traffic: 12_884_901_888,
	lifetime_used_traffic: 48_213_990_400,
	used_traffic_percent: 0.18,
	status: "active",
	last_login_ip: "127.0.0.1",
	telegram_id: 100000001,
	discord_id: "",
	email: "asli@example.com",
};

const iso = (offsetHours) =>
	new Date(Date.parse(now) - offsetHours * 3600_000).toISOString();

export const session = {
	state: "active",
	admin,
	permissions_version: "mock-1",
	totp_enabled: true,
	require_2fa: false,
};

export const users = [
	{
		username: "amir",
		status: "active",
		used_traffic: 8_912_345_678,
		lifetime_used_traffic: 92_345_678_901,
		download_speed: 4_718_592,
		upload_speed: 1_310_720,
		expire: iso(-24 * 30),
		online_at: iso(0.02),
		is_online: true,
		used_traffic_percent: 0.31,
		belong_to: { id: 1, username: "asli" },
	},
	{
		username: "bahareh",
		status: "active",
		used_traffic: 2_244_901_120,
		lifetime_used_traffic: 15_902_113_408,
		download_speed: 0,
		upload_speed: 0,
		expire: iso(24 * 21),
		online_at: iso(4),
		is_online: false,
		used_traffic_percent: 0.62,
		belong_to: { id: 1, username: "asli" },
	},
	{
		username: "carol",
		status: "limited",
		used_traffic: 1_073_741_824,
		lifetime_used_traffic: 1_073_741_824,
		download_speed: 0,
		upload_speed: 0,
		expire: iso(24 * 3),
		online_at: iso(48),
		is_online: false,
		used_traffic_percent: 1,
		belong_to: { id: 2, username: "bahareh" },
	},
	{
		username: "dariush",
		status: "on_hold",
		used_traffic: 0,
		lifetime_used_traffic: 4_294_967_296,
		download_speed: 0,
		upload_speed: 0,
		expire: iso(-24 * 2),
		online_at: iso(96),
		is_online: false,
		used_traffic_percent: 0,
		belong_to: { id: 1, username: "asli" },
	},
	{
		username: "ehsan",
		status: "expired",
		used_traffic: 322_122_547,
		lifetime_used_traffic: 12_884_901_888,
		download_speed: 0,
		upload_speed: 0,
		expire: iso(-24 * 9),
		online_at: iso(120),
		is_online: false,
		used_traffic_percent: 0.44,
		belong_to: { id: 1, username: "asli" },
	},
];

export const admins = [
	admin,
	{
		...admin,
		id: 2,
		username: "bahareh",
		is_sudo: false,
		role: "standard",
		used_traffic: 1_610_612_736,
		lifetime_used_traffic: 6_442_450_944,
		last_login_ip: "10.0.0.7",
	},
	{
		...admin,
		id: 3,
		username: "reseller-one",
		is_sudo: false,
		role: "reseller",
		used_traffic: 268_435_456,
		lifetime_used_traffic: 1_073_741_824,
		last_login_ip: "10.0.0.9",
	},
	{
		...admin,
		id: 4,
		username: "sudo-two",
		is_sudo: true,
		role: "sudo",
		used_traffic: 0,
		lifetime_used_traffic: 0,
		last_login_ip: "10.0.0.11",
	},
];

export const nodes = [
	{
		id: 1,
		name: "tehran-1",
		status: "connected",
		address: "10.20.0.11",
		port: 62050,
		is_connected: true,
		connected_at: iso(72),
		last_online: iso(0),
		version: "1.0.0",
		usage: {
			upload_speed: 1_572_864,
			download_speed: 12_582_912,
			used_traffic: 34_359_738_368,
			up: 68_719_476_736,
			down: 137_438_953_472,
		},
	},
	{
		id: 2,
		name: "isfahan-2",
		status: "disconnected",
		address: "10.20.0.12",
		port: 62050,
		is_connected: false,
		connected_at: iso(-240),
		last_online: iso(6),
		version: "1.0.0",
		usage: {
			upload_speed: 0,
			download_speed: 0,
			used_traffic: 12_884_901_888,
			up: 25_769_803_776,
			down: 51_539_607_552,
		},
	},
];

export const hosts = [
	{
		id: 1,
		host: "example.com",
		remark: "Main site",
		address: ["example.com"],
		port: 443,
		enable: true,
		up: 2_147_483_648,
		down: 5_368_709_120,
		client_user_id: 0,
		tls: "on",
	},
	{
		id: 2,
		host: "docs.example.com",
		remark: "Documentation",
		address: ["docs.example.com"],
		port: 8443,
		enable: true,
		up: 134_217_728,
		down: 402_653_184,
		client_user_id: 1,
		tls: "on",
	},
	{
		id: 3,
		host: "legacy.example.net",
		remark: "Retired",
		address: ["legacy.example.net"],
		port: 80,
		enable: false,
		up: 0,
		down: 0,
		client_user_id: 2,
		tls: "none",
	},
];

export const services = [
	{
		id: 1,
		name: "Traffic monitor",
		description: "Per-user traffic accounting",
		service_type: "TrafficMonitor",
	},
	{
		id: 2,
		name: "Admin bot",
		description: "Telegram operator for admins",
		service_type: "Server",
	},
	{
		id: 3,
		name: "Placeholder",
		description: "Placeholder service",
		service_type: "Plugin",
	},
];

export const haproxy = {
	config: "global\n  stats enable\n",
	templates: [],
	preview_url: "http://127.0.0.1:8080",
};

export const core = {
	version: "is.0.0.1",
	latest_version: "is.0.0.1",
	started: true,
	uptime: 864_000,
	platform: "linux",
	xray_version: "25.1.1",
	geo: { updated_at: iso(24), templates: [] },
	logs_websocket: null,
};

export const system = {
	uptime: 864_000,
	uptime_human: "1d",
	version: "is.0.0.1",
	commit: "b01fd18",
	start_time: iso(-24),
	core_version: core.version,
	xray_version: core.xray_version,
	cpu: 12.4,
	ram: 3_221_225_472,
	total_ram: 8_589_934_592,
	disk: 42_949_672_960,
	total_disk: 214_748_364_800,
	net_connections: 143,
	total_users: users.length,
	total_inbounds: 2,
	total_outbounds: 3,
	load: [0.42, 0.51, 0.47],
};

export const metrics = {
	labels: ["10:00", "11:00", "12:00", "13:00", "14:00", "15:00"],
	upload: [1_572_864, 1_048_576, 2_097_152, 3_145_728, 2_621_440, 1_835_008],
	download: [8_912_896, 12_582_912, 9_437_184, 15_728_640, 11_534_336, 7_340_032],
	cpu: [18, 24, 31, 27, 22, 19],
	ram: [38, 39, 41, 44, 42, 40],
};

export const inbounds = [
	{
		id: 1,
		tag: "inbound-1",
		port: 62050,
		protocol: "vless",
		listen: "0.0.0.0",
		remark: "Main inbound",
		enable: true,
		settings: { clients: [] },
		streamSettings: { network: "tcp", security: "reality" },
	},
	{
		id: 2,
		tag: "inbound-2",
		port: 62051,
		protocol: "vmess",
		listen: "0.0.0.0",
		remark: "Legacy inbound",
		enable: false,
		settings: { clients: [] },
		streamSettings: { network: "ws", security: "tls" },
	},
];

export const outbounds = [
	{
		id: 1,
		tag: "direct",
		protocol: "freedom",
		remark: "Direct",
		enable: true,
		settings: {},
		streamSettings: {},
	},
	{
		id: 2,
		tag: "blocked",
		protocol: "blackhole",
		remark: "Blocked",
		enable: true,
		settings: {},
		streamSettings: {},
	},
];

export const xray = {
	config: '{"log":{"loglevel":"warning"}}',
	settings: {
		log: { loglevel: "warning" },
		inbounds,
		outbounds,
		routing: { rules: [], balancers: [] },
	},
	version: core.xray_version,
};

export const settingsAll = {
	panel: {
		panel_name: "Gamaj",
		logo: "",
		theme: "dark",
		page_title: "Gamaj",
		timezone: "Asia/Tehran",
		admin_login_slug: "admin",
	},
	users: {
		default_traffic: 10_737_418_240,
		default_expire_days: 30,
		default_status: "active",
		users_filter: "",
	},
	security: {
		allow_insecure_tls: false,
		tls_fingerprint: "chrome",
	},
	subscriptions: {
		title: "Gamaj subscription",
		header: "Your subscription link",
		support_url: "",
		telegram_link: "",
		footer: "Coded by AsliCode",
	},
	integrations: {
		telegram_enabled: false,
		telegram_token: "",
		telegram_admin_id: "",
	},
};

export const placeholders = {
	title: "Under construction",
	description: "This page is being built.",
	button_text: "Back to dashboard",
	show_logo: true,
};

export const externalApps = [
	{
		id: 1,
		name: "MirzaBot",
		domain: "bot.example.com",
		link: "http://bot.example.com",
		installed: true,
		version: "2.4.1",
		template: "mirzabot",
		file_name: "bot.zip",
		static_cache_seconds: 3600,
		not_found_file: "404.html",
		admin_id: 1,
	},
	{
		id: 2,
		name: "FaXiMa panel",
		domain: "panel.example.com",
		link: "https://panel.example.com",
		installed: false,
		version: "",
		template: "faoxima",
		file_name: "panel.zip",
		static_cache_seconds: 3600,
		not_found_file: "404.html",
		admin_id: 2,
	},
];

export const recentActions = [
	{
		id: "a-1",
		action: "user.update",
		admin_username: "asli",
		created_at: iso(0.5),
		ip: "127.0.0.1",
		details: { username: "amir" },
	},
	{
		id: "a-2",
		action: "core.restart",
		admin_username: "asli",
		created_at: iso(2),
		ip: "127.0.0.1",
		details: {},
	},
	{
		id: "a-3",
		action: "admin.create",
		admin_username: "asli",
		created_at: iso(9),
		ip: "10.0.0.4",
		details: { username: "reseller-one" },
	},
];

export const apiKeys = [
	{
		id: 1,
		name: "Reporting key",
		key: "gm_live_0000_1111_2222_3333",
		created_at: iso(240),
		last_used_at: iso(2),
		usage_count: 1284,
	},
];

export const adminSessions = [
	{
		id: 1,
		created_at: iso(240),
		last_seen_at: iso(0),
		expires_at: iso(24 * 7),
		ip_address: "127.0.0.1",
		user_agent: "Mozilla/5.0",
		current: true,
	},
	{
		id: 2,
		created_at: iso(96),
		last_seen_at: iso(30),
		expires_at: iso(24 * 3),
		ip_address: "10.0.0.8",
		user_agent: "Mozilla/5.0",
		current: false,
	},
];

export const xrayLogs = [
	{ time: iso(0.1), level: "info", message: "Xray 25.1.1 started" },
	{ time: iso(0.2), level: "info", message: "Listening on 0.0.0.0:62050" },
	{ time: iso(0.3), level: "warning", message: "Reality handshake fallback" },
];

export const accessInsights = {
	total_requests: 128_442,
	unique_ips: 412,
	bandwidth: 987_654_321_000,
	top_paths: [
		{ path: "/api/v1/user/self", count: 54_321 },
		{ path: "/api/v1/user/sub", count: 22_104 },
		{ path: "/api/v1/user/info", count: 18_002 },
	],
	top_ips: [
		{ ip: "10.0.0.21", count: 9_112 },
		{ ip: "10.0.0.22", count: 8_004 },
	],
};

export const myAccount = {
	admin,
	usage: {
		used_traffic: admin.used_traffic,
		lifetime_used_traffic: admin.lifetime_used_traffic,
		download_speed: 1_048_576,
		upload_speed: 262_144,
	},
	limits: { concurrent: 0, cores: 0 },
};

export const warp = {
	registered: false,
	license: "",
	config: {},
};

export const nodeSettings = {
	enable_node: true,
	node_cert: "",
	node_cert_key: "",
	node_port: 62050,
	node_type: "standalone",
};

export const geoTemplates = [
	{ id: 1, name: "iran", type: "ip" },
	{ id: 2, name: "private", type: "ip" },
];
