export type UsageTone = "unlimited" | "ok" | "warn" | "critical";

export const getUsageTone = (
	percent: number,
	isUnlimited: boolean,
): UsageTone => {
	if (isUnlimited) return "unlimited";
	if (percent >= 85) return "critical";
	if (percent >= 65) return "warn";
	return "ok";
};

export const usageToneGradients: Record<UsageTone, string> = {
	unlimited: "linear-gradient(90deg, var(--gm-panel-accent), var(--gm-panel-accent))",
	ok: "linear-gradient(90deg, var(--gm-success), var(--gm-success))",
	warn: "linear-gradient(90deg, var(--gm-warning), var(--gm-warning))",
	critical: "linear-gradient(90deg, var(--gm-danger), var(--gm-danger))",
};
