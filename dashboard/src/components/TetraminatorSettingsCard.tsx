import { useEffect, useState } from "react";
import {
	Box,
	Button,
	Flex,
	Heading,
	Spinner,
	Stack,
	Switch,
	Text,
	Input,
	FormControl,
	FormLabel,
	useToast,
} from "@chakra-ui/react";
import {
	getSalesSettings,
	updateSalesSettings,
	type SalesSettingsResponse,
} from "../service/settings";
import { useTranslation } from "react-i18next";

/**
 * TetraminatorSettingsCard — sudo-only card that manages the Tetraminator
 * payment gateway used by Gamaj Bot sales. The API key is write-only: the
 * dashboard shows a masked value and the panel never echoes the secret.
 */
export const TetraminatorSettingsCard = () => {
	const { t } = useTranslation();
	const toast = useToast();
	const [settings, setSettings] = useState<SalesSettingsResponse | null>(null);
	const [loading, setLoading] = useState(true);
	const [saving, setSaving] = useState(false);
	const [baseUrl, setBaseUrl] = useState("");
	const [apiKey, setApiKey] = useState("");
	const [minPrice, setMinPrice] = useState("50000");
	const [maxPrice, setMaxPrice] = useState("10000000");
	const [callbackURL, setCallbackURL] = useState("");

	useEffect(() => {
		let mounted = true;
		getSalesSettings()
			.then((response) => {
				if (!mounted) return;
				setSettings(response);
				setBaseUrl(response.tetraminator_base_url || "");
				setMinPrice(String(response.tetraminator_min_price ?? 50000));
				setMaxPrice(String(response.tetraminator_max_price ?? 10000000));
				setCallbackURL(response.tetraminator_callback_url || "");
			})
			.catch(() => {
				if (mounted) {
					toast({
						title: t("error"),
						status: "error",
						duration: 3000,
					});
				}
			})
			.finally(() => {
				if (mounted) setLoading(false);
			});
		return () => {
			mounted = false;
		};
	}, [t, toast]);

	const handleSave = async () => {
		setSaving(true);
		try {
			const payload: Record<string, unknown> = {
				tetraminator_base_url: baseUrl,
				tetraminator_min_price: Number(minPrice) || 50000,
				tetraminator_max_price: Number(maxPrice) || 10000000,
				tetraminator_callback_url: callbackURL,
			};
			if (apiKey.trim() !== "") {
				payload.tetraminator_api_key = apiKey.trim();
			}
			const updated = await updateSalesSettings(payload);
			setSettings(updated);
			setApiKey("");
			toast({
				title: t("settings.saved"),
				status: "success",
				duration: 2500,
			});
		} catch {
			toast({ title: t("error"), status: "error", duration: 3000 });
		} finally {
			setSaving(false);
		}
	};

	const handleToggle = async (enabled: boolean) => {
		setSaving(true);
		try {
			const updated = await updateSalesSettings({
				tetraminator_enabled: enabled,
			});
			setSettings(updated);
		} catch {
			toast({ title: t("error"), status: "error", duration: 3000 });
		} finally {
			setSaving(false);
		}
	};

	if (loading) {
		return (
			<Flex align="center" justify="center" py={12}>
				<Spinner size="lg" />
			</Flex>
		);
	}

	const enabled = settings?.tetraminator_enabled ?? false;
	const configured = settings?.tetraminator_configured ?? false;

	return (
		<Stack spacing={4}>
			<Box className="master-settings-card" borderRadius="2xl" p={5}>
				<Flex justify="space-between" align="center" mb={4}>
					<Box>
						<Heading size="sm">
							{t("settings.sales.tetraminatorTitle")}
						</Heading>
						<Text fontSize="sm" color="gray.500" mt={1}>
							{t("settings.sales.tetraminatorDescription")}
						</Text>
					</Box>
					<Switch
						colorScheme="primary"
						isChecked={enabled}
						isDisabled={saving || !configured}
						onChange={(event) => handleToggle(event.target.checked)}
					/>
				</Flex>
				<Stack spacing={3}>
					<FormControl>
						<FormLabel fontSize="sm">
							{t("settings.sales.baseUrl")}
						</FormLabel>
						<Input
							value={baseUrl}
							onChange={(event) => setBaseUrl(event.target.value)}
							placeholder="https://pay.example.com"
						/>
					</FormControl>
					<FormControl>
						<FormLabel fontSize="sm">
							{t("settings.sales.apiKey")}
							{settings?.tetraminator_masked_key ? (
								<Text as="span" color="gray.500" ms={2}>
									({settings.tetraminator_masked_key})
								</Text>
							) : null}
						</FormLabel>
						<Input
							type="password"
							value={apiKey}
							onChange={(event) => setApiKey(event.target.value)}
							placeholder={
								configured
									? t("settings.sales.apiKeyKeep")
									: t("settings.sales.apiKeyPlaceholder")
							}
						/>
					</FormControl>
					<Flex gap={3}>
						<FormControl>
							<FormLabel fontSize="sm">
								{t("settings.sales.minPrice")}
							</FormLabel>
							<Input
								type="number"
								value={minPrice}
								onChange={(event) => setMinPrice(event.target.value)}
							/>
						</FormControl>
						<FormControl>
							<FormLabel fontSize="sm">
								{t("settings.sales.maxPrice")}
							</FormLabel>
							<Input
								type="number"
								value={maxPrice}
								onChange={(event) => setMaxPrice(event.target.value)}
							/>
						</FormControl>
					</Flex>
					<FormControl>
						<FormLabel fontSize="sm">
							{t("settings.sales.callbackURL")}
						</FormLabel>
						<Input
							value={callbackURL}
							onChange={(event) => setCallbackURL(event.target.value)}
							placeholder="https://panel.example.com"
						/>
					</FormControl>
					<Button
						alignSelf="flex-end"
						colorScheme="primary"
						onClick={handleSave}
						isLoading={saving}
					>
						{t("settings.save")}
					</Button>
				</Stack>
			</Box>
		</Stack>
	);
};
