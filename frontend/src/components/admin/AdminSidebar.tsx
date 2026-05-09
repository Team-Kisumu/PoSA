"use client";

import type { UserProfile } from "@/lib/api";
import { logout } from "@/lib/api";

type View = "overview" | "users" | "credentials" | "proofs" | "settings";

interface Props {
  view: View;
  setView: (v: View) => void;
  user: UserProfile;
  open: boolean;
  onClose: () => void;
}

export default function AdminSidebar({ view, setView, user, open, onClose }: Props) {
  const handleLogout = async () => { await logout(); window.location.href = "/"; };
  const handleNav = (v: View) => { setView(v); onClose(); };

  return (
    <>
      {/* Mobile overlay */}
      {open && <div className="fixed inset-0 z-40 bg-black/60 lg:hidden" onClick={onClose} />}

      <aside className={`fixed lg:sticky top-0 left-0 z-50 w-[260px] bg-[#0c1222] border-r border-white/[0.06] flex flex-col h-screen transition-transform duration-300 lg:translate-x-0 ${
        open ? "translate-x-0" : "-translate-x-full"
      }`}>
      {/* Brand */}
      <div className="px-6 py-5 border-b border-white/[0.06]">
        <div className="flex items-center gap-3">
          <img src="/appicon.png" alt="PoSA" className="w-10 h-10 rounded-xl" />
          <div>
            <p className="text-white font-semibold text-[15px] leading-tight">PoSA</p>
            <p className="text-[11px] text-zinc-500 leading-tight mt-0.5">Admin Panel</p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-3 py-5 overflow-y-auto">
        {/* Main Section */}
        <p className="px-3 mb-2 text-[10px] font-semibold uppercase tracking-wider text-zinc-600">Main</p>
        <NavItem active={view === "overview"} onClick={() => handleNav("overview")} icon={<IconGrid />} label="Dashboard" />
        <NavItem active={view === "users"} onClick={() => handleNav("users")} icon={<IconUsers />} label="Users" />
        <NavItem active={view === "credentials"} onClick={() => handleNav("credentials")} icon={<IconShield />} label="Credentials" />
        <NavItem active={view === "proofs"} onClick={() => handleNav("proofs")} icon={<IconCube />} label="Proofs" />

        {/* System Section */}
        <p className="px-3 mt-6 mb-2 text-[10px] font-semibold uppercase tracking-wider text-zinc-600">System</p>
        <NavItem active={view === "settings"} onClick={() => handleNav("settings")} icon={<IconGear />} label="Settings" />
        <NavLink href="/analyze" icon={<IconBeaker />} label="Analyze Code" />
        <NavLink href="/" icon={<IconGlobe />} label="View Site" />
        <NavLink href="https://github.com/Team-Kisumu/PoSA" icon={<IconCode />} label="GitHub" external />
      </nav>

      {/* User Card */}
      <div className="px-4 py-4 border-t border-white/[0.06]">
        <div className="flex items-center gap-3 px-2">
          <img src={user.avatar_url} alt="" className="w-9 h-9 rounded-full ring-2 ring-emerald-500/20" />
          <div className="flex-1 min-w-0">
            <p className="text-[13px] text-white font-medium truncate">{user.username}</p>
            <p className="text-[11px] text-emerald-400">Admin</p>
          </div>
        </div>
        <button
          onClick={handleLogout}
          className="w-full mt-3 flex items-center gap-2 px-3 py-2 text-[12px] text-zinc-500 hover:text-red-400 rounded-lg hover:bg-red-500/5 transition-all"
        >
          <IconLogout />
          Sign Out
        </button>
      </div>
      </aside>
    </>
  );
}

// Nav item (button)
function NavItem({ active, onClick, icon, label }: { active: boolean; onClick: () => void; icon: React.ReactNode; label: string }) {
  return (
    <button
      onClick={onClick}
      className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-[13px] font-medium transition-all mb-0.5 ${
        active
          ? "bg-emerald-500/10 text-emerald-400"
          : "text-zinc-400 hover:text-zinc-200 hover:bg-white/[0.03]"
      }`}
    >
      <span className={active ? "text-emerald-400" : "text-zinc-500"}>{icon}</span>
      {label}
      {active && <span className="ml-auto w-1.5 h-1.5 rounded-full bg-emerald-400" />}
    </button>
  );
}

// Nav link (anchor)
function NavLink({ href, icon, label, external }: { href: string; icon: React.ReactNode; label: string; external?: boolean }) {
  return (
    <a
      href={href}
      target={external ? "_blank" : undefined}
      rel={external ? "noopener noreferrer" : undefined}
      className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-[13px] font-medium text-zinc-400 hover:text-zinc-200 hover:bg-white/[0.03] transition-all mb-0.5"
    >
      <span className="text-zinc-500">{icon}</span>
      {label}
      {external && <span className="ml-auto text-zinc-600 text-[10px]">&#8599;</span>}
    </a>
  );
}

// Icons (inline SVG for zero dependencies)
function IconGrid() { return <svg className="w-[18px] h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z" /></svg>; }
function IconUsers() { return <svg className="w-[18px] h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" /></svg>; }
function IconShield() { return <svg className="w-[18px] h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" /></svg>; }
function IconCube() { return <svg className="w-[18px] h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M21 7.5l-9-5.25L3 7.5m18 0l-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" /></svg>; }
function IconGear() { return <svg className="w-[18px] h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z" /><path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>; }
function IconGlobe() { return <svg className="w-[18px] h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0121 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0112 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 013 12c0-1.605.42-3.113 1.157-4.418" /></svg>; }
function IconCode() { return <svg className="w-[18px] h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M17.25 6.75L22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3l-4.5 16.5" /></svg>; }
function IconBeaker() { return <svg className="w-[18px] h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M9.75 3.104v5.714a2.25 2.25 0 01-.659 1.591L5 14.5M9.75 3.104c-.251.023-.501.05-.75.082m.75-.082a24.301 24.301 0 014.5 0m0 0v5.714c0 .597.237 1.17.659 1.591L19.8 15.3M14.25 3.104c.251.023.501.05.75.082M19.8 15.3l-1.57.393A9.065 9.065 0 0112 15a9.065 9.065 0 00-6.23.693L5 14.5m14.8.8l1.402 1.402c1.232 1.232.65 3.318-1.067 3.611A48.309 48.309 0 0112 21c-2.773 0-5.491-.235-8.135-.687-1.718-.293-2.3-2.379-1.067-3.61L5 14.5" /></svg>; }
function IconLogout() { return <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15m3 0l3-3m0 0l-3-3m3 3H9" /></svg>; }

export type { View };
