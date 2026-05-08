"use client";

import { useEffect, useState } from "react";
import { getMe, logout } from "@/lib/api";
import type { UserProfile } from "@/lib/api";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

type View = "overview" | "users" | "credentials";

interface Stats {
  total_users: number;
  total_submissions: number;
}

interface User {
  id: number;
  github_id: number;
  username: string;
  avatar_url: string;
  email: string;
  role: string;
  created_at: string;
}

interface Credential {
  id: number;
  user_id: number;
  type: string;
  name: string;
  score: number;
  cid: string;
  tx_hash: string;
  created_at: string;
}

export default function AdminPage() {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [view, setView] = useState<View>("overview");
  const [stats, setStats] = useState<Stats | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [credentials, setCredentials] = useState<Credential[]>([]);
  const [userTotal, setUserTotal] = useState(0);
  const [credTotal, setCredTotal] = useState(0);

  useEffect(() => {
    getMe().then((u) => { setUser(u); setLoading(false); });
  }, []);

  useEffect(() => {
    if (!user || user.role !== "admin") return;
    fetch(`${API_URL}/api/admin/stats`, { credentials: "include" })
      .then((r) => r.json()).then((d) => setStats(d.data));
    fetch(`${API_URL}/api/admin/users?limit=50`, { credentials: "include" })
      .then((r) => r.json()).then((d) => { setUsers(d.data?.users || []); setUserTotal(d.data?.total || 0); });
    fetch(`${API_URL}/api/admin/credentials?limit=50`, { credentials: "include" })
      .then((r) => r.json()).then((d) => { setCredentials(d.data?.credentials || []); setCredTotal(d.data?.total || 0); });
  }, [user]);

  const updateRole = async (id: number, role: string) => {
    await fetch(`${API_URL}/api/admin/users/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ role }),
      credentials: "include",
    });
    setUsers((prev) => prev.map((u) => (u.id === id ? { ...u, role } : u)));
  };

  const handleLogout = async () => { await logout(); window.location.href = "/"; };

  if (loading) {
    return (
      <div className="min-h-screen bg-[#0a0e1a] flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (!user || user.role !== "admin") {
    return (
      <div className="min-h-screen bg-[#0a0e1a] flex items-center justify-center">
        <div className="text-center">
          <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-red-500/10 flex items-center justify-center">
            <svg className="w-8 h-8 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" /></svg>
          </div>
          <p className="text-xl font-bold text-white mb-2">Access Denied</p>
          <p className="text-zinc-500 text-sm">Admin privileges required.</p>
          <a href="/" className="inline-block mt-4 text-sm text-emerald-400 hover:underline">Back to Home</a>
        </div>
      </div>
    );
  }

  const navItems = [
    { id: "overview" as View, label: "Overview", icon: <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" /></svg> },
    { id: "users" as View, label: "Users", icon: <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" /></svg> },
    { id: "credentials" as View, label: "Credentials", icon: <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg> },
  ];

  return (
    <div className="min-h-screen bg-[#0a0e1a] flex">
      {/* Sidebar */}
      <aside className="w-64 bg-[#0f1425] border-r border-white/5 flex flex-col">
        <div className="p-6 border-b border-white/5">
          <div className="flex items-center gap-3">
            <img src="/appicon.png" alt="PoSA" className="w-9 h-9 rounded-lg" />
            <div>
              <p className="text-white font-bold text-sm">PoSA Admin</p>
              <p className="text-zinc-500 text-xs">Control Panel</p>
            </div>
          </div>
        </div>

        <nav className="flex-1 p-4 space-y-1">
          {navItems.map((item) => (
            <button
              key={item.id}
              onClick={() => setView(item.id)}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-all ${
                view === item.id
                  ? "bg-emerald-500/10 text-emerald-400 shadow-lg shadow-emerald-500/5"
                  : "text-zinc-400 hover:text-white hover:bg-white/5"
              }`}
            >
              {item.icon}
              {item.label}
            </button>
          ))}
        </nav>

        <div className="p-4 border-t border-white/5">
          <div className="flex items-center gap-3 px-4 py-3">
            <img src={user.avatar_url} alt="" className="w-9 h-9 rounded-full ring-2 ring-emerald-500/30" />
            <div className="flex-1 min-w-0">
              <p className="text-white text-sm font-medium truncate">{user.username}</p>
              <p className="text-emerald-400 text-xs">Admin</p>
            </div>
          </div>
          <button onClick={handleLogout} className="w-full mt-2 px-4 py-2 text-xs text-zinc-500 hover:text-red-400 transition-colors text-left">
            Sign Out
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-auto">
        {/* Header */}
        <header className="sticky top-0 z-10 bg-[#0a0e1a]/80 backdrop-blur-xl border-b border-white/5 px-8 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-xl font-bold text-white">
                {view === "overview" && "Dashboard Overview"}
                {view === "users" && "User Management"}
                {view === "credentials" && "Credentials"}
              </h1>
              <p className="text-zinc-500 text-sm mt-0.5">
                {new Date().toLocaleDateString(undefined, { weekday: "long", year: "numeric", month: "long", day: "numeric" })}
              </p>
            </div>
            <a href="/" className="text-xs text-zinc-500 hover:text-white transition-colors px-3 py-1.5 rounded-lg border border-white/10 hover:border-white/20">
              View Site
            </a>
          </div>
        </header>

        <div className="p-8">
          {/* Overview */}
          {view === "overview" && (
            <div className="space-y-8">
              {/* Stat Cards */}
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
                <StatCard label="Total Users" value={stats?.total_users ?? 0} icon={<svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" /></svg>} color="emerald" />
                <StatCard label="Total Submissions" value={stats?.total_submissions ?? 0} icon={<svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>} color="blue" />
                <StatCard label="Admin Users" value={users.filter((u) => u.role === "admin").length} icon={<svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>} color="purple" />
                <StatCard label="Avg Score" value={credentials.length > 0 ? Math.round(credentials.reduce((a, c) => a + c.score, 0) / credentials.length) : 0} icon={<svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" /></svg>} color="amber" suffix="/100" />
              </div>

              {/* Recent Users */}
              <div className="bg-[#0f1425] border border-white/5 rounded-2xl overflow-hidden">
                <div className="px-6 py-4 border-b border-white/5 flex items-center justify-between">
                  <h3 className="text-white font-semibold">Recent Users</h3>
                  <button onClick={() => setView("users")} className="text-xs text-emerald-400 hover:underline">View All</button>
                </div>
                <div className="divide-y divide-white/5">
                  {users.slice(0, 5).map((u) => (
                    <div key={u.id} className="px-6 py-4 flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <img src={u.avatar_url} alt="" className="w-9 h-9 rounded-full" />
                        <div>
                          <p className="text-white text-sm font-medium">{u.username}</p>
                          <p className="text-zinc-500 text-xs">{u.email || "No email"}</p>
                        </div>
                      </div>
                      <span className={`px-2.5 py-1 text-xs rounded-full ${
                        u.role === "admin" ? "bg-emerald-500/10 text-emerald-400" : "bg-white/5 text-zinc-400"
                      }`}>{u.role}</span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Recent Credentials */}
              <div className="bg-[#0f1425] border border-white/5 rounded-2xl overflow-hidden">
                <div className="px-6 py-4 border-b border-white/5 flex items-center justify-between">
                  <h3 className="text-white font-semibold">Recent Submissions</h3>
                  <button onClick={() => setView("credentials")} className="text-xs text-emerald-400 hover:underline">View All</button>
                </div>
                <div className="divide-y divide-white/5">
                  {credentials.slice(0, 5).map((c) => (
                    <div key={c.id} className="px-6 py-4 flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div className={`w-9 h-9 rounded-lg flex items-center justify-center text-sm font-bold ${
                          c.score >= 80 ? "bg-emerald-500/10 text-emerald-400" : c.score >= 50 ? "bg-amber-500/10 text-amber-400" : "bg-red-500/10 text-red-400"
                        }`}>{c.score}</div>
                        <div>
                          <p className="text-white text-sm font-medium truncate max-w-64">{c.name}</p>
                          <p className="text-zinc-500 text-xs">{c.type} &middot; {new Date(c.created_at).toLocaleDateString()}</p>
                        </div>
                      </div>
                      {c.cid && <a href={`https://gateway.lighthouse.storage/ipfs/${c.cid}`} target="_blank" rel="noopener noreferrer" className="text-xs text-blue-400 hover:underline">IPFS</a>}
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* Users View */}
          {view === "users" && (
            <div className="bg-[#0f1425] border border-white/5 rounded-2xl overflow-hidden">
              <div className="px-6 py-4 border-b border-white/5 flex items-center justify-between">
                <h3 className="text-white font-semibold">{userTotal} Users</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-white/5">
                      <th className="text-left px-6 py-3 text-zinc-500 font-medium">User</th>
                      <th className="text-left px-6 py-3 text-zinc-500 font-medium">Email</th>
                      <th className="text-left px-6 py-3 text-zinc-500 font-medium">Role</th>
                      <th className="text-left px-6 py-3 text-zinc-500 font-medium">Joined</th>
                      <th className="text-left px-6 py-3 text-zinc-500 font-medium">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-white/5">
                    {users.map((u) => (
                      <tr key={u.id} className="hover:bg-white/[0.02] transition-colors">
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-3">
                            <img src={u.avatar_url} alt="" className="w-8 h-8 rounded-full" />
                            <span className="text-white font-medium">{u.username}</span>
                          </div>
                        </td>
                        <td className="px-6 py-4 text-zinc-400">{u.email || "-"}</td>
                        <td className="px-6 py-4">
                          <span className={`px-2.5 py-1 text-xs rounded-full ${
                            u.role === "admin" ? "bg-emerald-500/10 text-emerald-400" : "bg-white/5 text-zinc-400"
                          }`}>{u.role}</span>
                        </td>
                        <td className="px-6 py-4 text-zinc-500">{new Date(u.created_at).toLocaleDateString()}</td>
                        <td className="px-6 py-4">
                          {u.role === "user" ? (
                            <button onClick={() => updateRole(u.id, "admin")} className="text-xs text-emerald-400 hover:text-emerald-300 transition-colors">Promote</button>
                          ) : (
                            <button onClick={() => updateRole(u.id, "user")} className="text-xs text-red-400 hover:text-red-300 transition-colors">Demote</button>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Credentials View */}
          {view === "credentials" && (
            <div className="bg-[#0f1425] border border-white/5 rounded-2xl overflow-hidden">
              <div className="px-6 py-4 border-b border-white/5 flex items-center justify-between">
                <h3 className="text-white font-semibold">{credTotal} Credentials</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-white/5">
                      <th className="text-left px-6 py-3 text-zinc-500 font-medium">ID</th>
                      <th className="text-left px-6 py-3 text-zinc-500 font-medium">Name</th>
                      <th className="text-left px-6 py-3 text-zinc-500 font-medium">Type</th>
                      <th className="text-left px-6 py-3 text-zinc-500 font-medium">Score</th>
                      <th className="text-left px-6 py-3 text-zinc-500 font-medium">CID</th>
                      <th className="text-left px-6 py-3 text-zinc-500 font-medium">Date</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-white/5">
                    {credentials.map((c) => (
                      <tr key={c.id} className="hover:bg-white/[0.02] transition-colors">
                        <td className="px-6 py-4 text-zinc-400 font-mono">#{c.id}</td>
                        <td className="px-6 py-4 text-white max-w-48 truncate">{c.name}</td>
                        <td className="px-6 py-4">
                          <span className="px-2.5 py-1 text-xs rounded-full bg-white/5 text-zinc-400">{c.type}</span>
                        </td>
                        <td className="px-6 py-4">
                          <span className={`font-bold ${c.score >= 80 ? "text-emerald-400" : c.score >= 50 ? "text-amber-400" : "text-red-400"}`}>{c.score}</span>
                        </td>
                        <td className="px-6 py-4">
                          {c.cid ? (
                            <a href={`https://gateway.lighthouse.storage/ipfs/${c.cid}`} target="_blank" rel="noopener noreferrer" className="text-blue-400 hover:underline font-mono text-xs">{c.cid.slice(0, 16)}...</a>
                          ) : <span className="text-zinc-600">-</span>}
                        </td>
                        <td className="px-6 py-4 text-zinc-500">{new Date(c.created_at).toLocaleDateString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

// Stat Card Component
function StatCard({ label, value, icon, color, suffix = "" }: { label: string; value: number; icon: React.ReactNode; color: string; suffix?: string }) {
  const colors: Record<string, string> = {
    emerald: "bg-emerald-500/10 text-emerald-400",
    blue: "bg-blue-500/10 text-blue-400",
    purple: "bg-purple-500/10 text-purple-400",
    amber: "bg-amber-500/10 text-amber-400",
  };
  return (
    <div className="bg-[#0f1425] border border-white/5 rounded-2xl p-6">
      <div className={`w-12 h-12 rounded-xl flex items-center justify-center mb-4 ${colors[color]}`}>
        {icon}
      </div>
      <p className="text-3xl font-bold text-white">{value}{suffix}</p>
      <p className="text-zinc-500 text-sm mt-1">{label}</p>
    </div>
  );
}
