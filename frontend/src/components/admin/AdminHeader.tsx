"use client";

import type { View } from "./AdminSidebar";

interface Props {
  view: View;
  onMenuToggle: () => void;
}

const viewTitles: Record<View, { title: string; subtitle: string }> = {
  overview: { title: "Dashboard", subtitle: "Platform overview and key metrics" },
  users: { title: "User Management", subtitle: "Manage registered users and roles" },
  credentials: { title: "Credentials", subtitle: "All minted proof-of-skill credentials" },
  proofs: { title: "On-Chain Proofs", subtitle: "Blockchain-anchored evaluations" },
  settings: { title: "Settings", subtitle: "System configuration" },
};

export default function AdminHeader({ view, onMenuToggle }: Props) {
  const { title, subtitle } = viewTitles[view];
  const now = new Date();
  const dateStr = now.toLocaleDateString(undefined, { weekday: "long", month: "short", day: "numeric", year: "numeric" });

  return (
    <header className="sticky top-0 z-10 bg-[#080d19]/90 backdrop-blur-xl border-b border-white/[0.06] px-4 sm:px-8 py-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button onClick={onMenuToggle} className="lg:hidden p-2 -ml-2 text-zinc-400 hover:text-white rounded-lg hover:bg-white/5 transition-colors">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" /></svg>
          </button>
          <div>
            <h1 className="text-[16px] sm:text-[18px] font-semibold text-white">{title}</h1>
            <p className="text-[11px] sm:text-[12px] text-zinc-500 mt-0.5 hidden sm:block">{subtitle}</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-[11px] sm:text-[12px] text-zinc-500 hidden md:block">{dateStr}</span>
          <div className="flex items-center gap-1.5 px-2 sm:px-3 py-1.5 rounded-lg bg-emerald-500/10 border border-emerald-500/20">
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
            <span className="text-[10px] sm:text-[11px] text-emerald-400 font-medium hidden sm:inline">Online</span>
          </div>
        </div>
      </div>
    </header>
  );
}
