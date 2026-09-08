"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { LayoutGrid, Home, BarChart3, BookOpen } from "lucide-react";
import { cn } from "@/lib/utils";

const navItems = [
  {
    href: "/jurnal",
    label: "Jurnal",
    icon: LayoutGrid,
  },
  {
    href: "/aset",
    label: "Aset",
    icon: Home,
  },
  {
    href: "/laporan",
    label: "Laporan",
    icon: BarChart3,
  },
  {
    href: "/akun",
    label: "Akun",
    icon: BookOpen,
  },
];

export function BottomNav() {
  const pathname = usePathname();

  return (
    <div className="fixed bottom-0 left-1/2 -translate-x-1/2 w-full max-w-97.5 z-40">
      {/* blur backdrop */}
      <div className="absolute inset-0 bg-white/80 backdrop-blur-md border-t border-black/6" />

      <nav className="relative flex items-stretch safe-bottom">
        {navItems.map(({ href, label, icon: Icon }) => {
          const active = pathname.startsWith(href);
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex-1 flex flex-col items-center justify-center gap-1 py-3 no-tap transition-all duration-150",
                active
                  ? "text-primary-600"
                  : "text-secondary-400 hover:text-secondary-600",
              )}
            >
              <div className="relative">
                {active && (
                  <span className="absolute -inset-2 rounded-xl bg-primary-50 -z-10" />
                )}
                <Icon
                  size={20}
                  strokeWidth={active ? 2.2 : 1.8}
                  className="transition-transform duration-150 active:scale-90"
                />
              </div>
              <span
                className={cn(
                  "text-[10px] font-semibold tracking-wide transition-all",
                  active ? "text-primary-600" : "text-secondary-400",
                )}
              >
                {label}
              </span>
            </Link>
          );
        })}
      </nav>
    </div>
  );
}
