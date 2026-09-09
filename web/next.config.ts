import path from "node:path";
import createNextIntlPlugin from "next-intl/plugin";

const withNextIntl = createNextIntlPlugin();

const nextConfig = {
	turbopack: {
		root: path.resolve(__dirname, ".."),
	},
};

export default withNextIntl(nextConfig);
