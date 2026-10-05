import {
	Box,
	HStack,
	Image,
	Text,
	Tooltip,
	VStack,
} from "@chakra-ui/react";
import { useTranslation } from "react-i18next";
import IrancellLogo from "../assets/operators/irancell-svgrepo-com.svg";
import MCILogo from "../assets/operators/mci-svgrepo-com.svg";
import RightelLogo from "../assets/operators/rightel-svgrepo-com.svg";
import TCILogo from "../assets/operators/tci-svgrepo-com.svg";

type OperatorBrand = {
	keywords: string[];
	logo: string;
	background: string;
};

const operatorBrands: OperatorBrand[] = [
	{
		keywords: ["irancell", "iran cell", "mtn"],
		logo: IrancellLogo,
		background: "#ffd600",
	},
	{
		keywords: ["hamrah aval", "mci", "mobile communication company"],
		logo: MCILogo,
		background: "#ffffff",
	},
	{ keywords: ["rightel"], logo: RightelLogo, background: "#ffffff" },
	{
		keywords: ["mokhaberat", "tci", "iran telecommunication company"],
		logo: TCILogo,
		background: "#ffffff",
	},
];

// An operator with no brand of its own gets one identity for all of them: a
// neutral inset chip with the standard text. The old code gave each unknown
// operator a different hue, which is exactly the decoration the identity
// forbids — there is one mark style here, not five.
const operatorFallback = {
	bg: "panel.inset",
	color: "panel.text",
} as const;

const findOperatorBrand = (shortName?: string, owner?: string) => {
	const identity = `${shortName || ""} ${owner || ""}`.toLowerCase();
	return operatorBrands.find((brand) =>
		brand.keywords.some((keyword) => identity.includes(keyword)),
	);
};

const operatorInitials = (label: string) => {
	const parts = label.trim().split(/\s+/).filter(Boolean);
	if (parts.length > 1) return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
	return label.slice(0, 2).toUpperCase();
};

export const OperatorIdentity = ({
	shortName,
	owner,
	compact = false,
}: {
	shortName?: string;
	owner?: string;
	compact?: boolean;
}) => {
	const { t } = useTranslation();
	const label =
		shortName || owner || t("usersTable.operatorUnknown");
	const brand = findOperatorBrand(shortName, owner);
	const markSize = compact ? "24px" : "32px";
	const fallbackBg = operatorFallback.bg;
	const fallbackColor = operatorFallback.color;
	const identity = (
		<HStack spacing={compact ? 1.5 : 2} minW={0} align="center">
			<Box
				w={markSize}
				h={markSize}
				flexShrink={0}
				display="flex"
				alignItems="center"
				justifyContent="center"
				borderWidth="1px"
				borderColor="panel.border"
				borderRadius="md"
				bg={brand?.background || fallbackBg}
				color={brand ? undefined : fallbackColor}
				overflow="hidden"
			>
				{brand ? (
					<Image
						src={brand.logo}
						alt=""
						boxSize={compact ? "18px" : "24px"}
						objectFit="contain"
					/>
				) : (
					<Text
						fontSize={compact ? "2xs" : "xs"}
						fontWeight="800"
						lineHeight="1"
					>
						{operatorInitials(label)}
					</Text>
				)}
			</Box>
			<VStack spacing={0} align="start" minW={0} textAlign="start">
				<Text
					fontSize={compact ? "xs" : "sm"}
					fontWeight="semibold"
					noOfLines={1}
					maxW="full"
				>
					{label}
				</Text>
				{!compact && owner && owner !== shortName ? (
					<Text fontSize="xs" color="panel.textMuted" noOfLines={1} maxW="full">
						{owner}
					</Text>
				) : null}
			</VStack>
		</HStack>
	);

	return compact && owner && owner !== shortName ? (
		<Tooltip label={owner} hasArrow>
			{identity}
		</Tooltip>
	) : (
		identity
	);
};
