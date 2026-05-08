"use client";

import type { View } from "./AdminSidebar";

interface Props {
  view: View;
}

const viewTitles: Record<View, { title: string; subtitle: string }> = {
  overview: { title: "Dashboard", subtitle: "Platform overview and key metrics" },
  users: { title: "User Management", subtitle: "Manage registered users and roles" },
  credentials: { title: "Credentials", subtitle: "All minted proof-of-skill credentials" },
  proofs: { title: "On-Chain Proofs", subtitle: "Blockchain-anchored evaluations" },
  settings: { title: "Settings", subtitle: "System configuration" },
};

export default function AdminHeader({ view }: Props) {
  const { title, subtitle } = viewTitles[view];
  const now = new Date();
  const dateStr = now.toLocaleDateString(undefined, { weekday: "long", month: "short", day: "numeric", year: "numeric" });

  return (
    <header className="sticky top-0 z-10 bg-[#080d19]/90 backdrop-blur-xl border-b border-white/[0.06] px-8 py-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-[18px] font-semibold text-white">{title}</h1>
          <p className="text-[12px] text-zinc-500 mt-0.5">{subtitle}</p>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-[12px] text-zinc-500">{dateStr}</span>
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-emerald-500/10 border border-emerald-500/20">
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
            <span className="text-[11px] text-emerald-400 font-medium">System Online</span>
          </div>
        </div>
      </div>
    </header>
  );
}
