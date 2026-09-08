export type ColorShade =
	| 50
	| 100
	| 200
	| 300
	| 400
	| 500
	| 600
	| 700
	| 800
	| 900;
export type SemanticColor =
	| "primary"
	| "secondary"
	| "success"
	| "warning"
	| "info"
	| "danger";

type ShadeMap = Record<ColorShade, string> & { DEFAULT: string };

export const colors: Record<SemanticColor, ShadeMap> = {
	primary: {
		50: "#eef1fd",
		100: "#d5dcfa",
		200: "#adb8f5",
		300: "#8595f0",
		400: "#6172ea",
		500: "#3b4fd4",
		600: "#2f3fb8",
		700: "#23309a",
		800: "#18237c",
		900: "#0e165f",
		DEFAULT: "#3b4fd4",
	},
	secondary: {
		50: "#f5f4f0",
		100: "#e8e7e2",
		200: "#d0cec8",
		300: "#b4b2aa",
		400: "#8a8880",
		500: "#635f58",
		600: "#4a4840",
		700: "#36342e",
		800: "#26241f",
		900: "#1a1a2e",
		DEFAULT: "#635f58",
	},
	success: {
		50: "#e3f5ec",
		100: "#c2e9d6",
		200: "#86d4ad",
		300: "#4dbe88",
		400: "#27a86b",
		500: "#1a7a52",
		600: "#146044",
		700: "#0e4a34",
		800: "#093524",
		900: "#042115",
		DEFAULT: "#1a7a52",
	},
	warning: {
		50: "#fef3db",
		100: "#fde3a8",
		200: "#fbcf6a",
		300: "#f7b730",
		400: "#e09d14",
		500: "#c2840a",
		600: "#a06808",
		700: "#7c4e06",
		800: "#5a3804",
		900: "#3a2302",
		DEFAULT: "#c2840a",
	},
	info: {
		50: "#e6f1fb",
		100: "#c0d9f5",
		200: "#8dbced",
		300: "#5a9ee4",
		400: "#3382d8",
		500: "#1a66c4",
		600: "#1451a0",
		700: "#0e3e7d",
		800: "#082c5b",
		900: "#041b3a",
		DEFAULT: "#1a66c4",
	},
	danger: {
		50: "#fcecea",
		100: "#f8cdc9",
		200: "#f19c95",
		300: "#e86b60",
		400: "#dc4133",
		500: "#c0392b",
		600: "#9e2a20",
		700: "#7b1e17",
		800: "#58130f",
		900: "#360b09",
		DEFAULT: "#c0392b",
	},
};

export const accountingColors = {
	debit: { bg: "bg-danger-50", text: "text-danger-700", dot: "bg-danger-500" },
	credit: {
		bg: "bg-success-50",
		text: "text-success-700",
		dot: "bg-success-500",
	},
} as const;
