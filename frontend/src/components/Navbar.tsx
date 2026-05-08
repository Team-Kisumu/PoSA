"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { getMe, getLoginUrl, logout } from "@/lib/api";
import type { UserProfile } from "@/lib/api";

export default function Navbar() {
  const pathname = usePathname();
  const [user, setUser] = useState<UserProfile | null>(null);

  useEffect(() => {
    getMe().then(setUser);
  }, [pathname]);

  const handleLogout = async () => {
    await logout();
    setUser(null);
    window.location.href = "/";
  };

  const linkClass = (href: string) =>
    `text-sm font-medium transition-colors ${
      pathname === href
        ? "text-black dark:text-white"
        : "text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-300"
    }`;

  return (
    <header className="border-b border-zinc-200 dark:border-zinc-800 bg-white/80 dark:bg-black/80 backdrop-blur-sm sticky top-0 z-50">
      <div className="max-w-6xl mx-auto px-6 py-4 flex items-center justify-between">
        <Link href="/" className="group flex items-center gap-3">
          <img src="/appicon.png" alt="PoSA" className="w-8 h-8 rounded-lg" />
          <div>
            <span className="text-lg font-bold text-black dark:text-white group-hover:text-zinc-600 dark:group-hover:text-zinc-300 transition-colors">
              PoSA
            </span>
            <span className="hidden sm:inline text-xs text-zinc-400 ml-2">Proof-of-Skill AI</span>
          </div>
        </Link>
        <nav className="flex items-center gap-4 sm:gap-6">
          <Link href="/analyze" className={linkClass("/analyze")}>Analyze</Link>
          <Link href="/proofs" className={linkClass("/proofs")}>Proofs</Link>
          <Link href="/verify" className={linkClass("/verify")}>Verify</Link>
          {user?.role === "admin" && (
            <Link href="/admin" className={linkClass("/admin")}>Admin</Link>
          )}
          <a
            href="https://github.com/Team-Kisumu/PoSA"
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-300 transition-colors"
          >
            GitHub
          </a>
          {user ? (
            <div className="flex items-center gap-3">
              <img src={user.avatar_url} alt={user.username} className="w-7 h-7 rounded-full" />
              <button onClick={handleLogout} className="text-sm text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-300 transition-colors">
                Logout
              </button>
            </div>
          ) : (
            <a
              href={getLoginUrl()}
              className="text-sm font-medium text-white bg-black dark:bg-white dark:text-black px-3 py-1.5 rounded-lg hover:bg-zinc-800 dark:hover:bg-zinc-200 transition-colors"
            >
              Sign in
            </a>
          )}
        </nav>
      </div>
    </header>
  );
}
