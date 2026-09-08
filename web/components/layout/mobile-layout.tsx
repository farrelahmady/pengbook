import { BottomNav } from "./bottom-nav";

interface MobileLayoutProps {
  children: React.ReactNode;
}

export function MobileLayout({ children }: MobileLayoutProps) {
  return (
    <div className="phone-container bg-secondary-50">
      <main className="pb-18">{children}</main>
      <BottomNav />
    </div>
  );
}
