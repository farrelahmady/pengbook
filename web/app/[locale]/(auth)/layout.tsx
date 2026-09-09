import { BottomNav } from "@/components/layout/bottom-nav";

export default function AuthLayout({
	children,
}: {
	children: React.ReactNode;
}) {
	return (
		<>
			{children}
			<BottomNav />
		</>
	);
}
