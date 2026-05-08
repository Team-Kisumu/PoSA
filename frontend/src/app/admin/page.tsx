"use client";

import { useEffect, useState, useCallback } from "react";
import { getMe } from "@/lib/api";
import type { UserProfile } from "@/lib/api";
import AdminSidebar from "@/components/admin/AdminSidebar";
import AdminHeader from "@/components/admin/AdminHeader";
import StatCard from "@/components/admin/StatCard";
import type { View } from "@/components/admin/AdminSidebar";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

interface Stats { total_users: number; total_submissions: number }
interface User { id: number; github_id: number; username: string; avatar_url: string; email: string; role: string; created_at: string }
interface Credential { id: number; user_id: number; type: string; name: string; score: number; cid: string; tx_hash: string; created_at: string }

export default function AdminPage() {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [authLoading, setAuthLoading] = useState(true);
  const [view, setView] = useState<View>("overview");
  const [stats, setStats] = useState<Stats | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [credentials, setCredentials] = useState<Credential[]>([]);
  const [dataLoading, setDataLoading] = useState(true);
  const [error, setError] = useState("");
  const [userTotal, setUserTotal] = useState(0);
  const [credTotal, setCredTotal] = useState(0);

  useEffect(() => { getMe().then((u) => { setUser(u); setAuthLoading(false); }); }, []);

  const fetchData = useCallback(async () => {
    if (!user || user.role !== "admin") return;
    setDataLoading(true);
    setError("");
    try {
      const [statsRes, usersRes, credsRes] = await Promise.all([
        fetch(`${API_URL}/api/admin/stats`, { credentials: "include" }),
        fetch(`${API_URL}/api/admin/users?limit=50`, { credentials: "include" }),
        fetch(`${API_URL}/api/admin/credentials?limit=50`, { credentials: "include" }),
      ]);
      if (!statsRes.ok || !usersRes.ok || !credsRes.ok) throw new Error("Failed to fetch data");
      const [statsData, usersData, credsData] = await Promise.all([statsRes.json(), usersRes.json(), credsRes.json()]);
      setStats(statsData.data);
      setUsers(usersData.data?.users || []);
      setUserTotal(usersData.data?.total || 0);
      setCredentials(credsData.data?.credentials || []);
      setCredTotal(credsData.data?.total || 0);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load dashboard data");
    } finally {
      setDataLoading(false);
    }
  }, [user]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const updateRole = async (id: number, role: string) => {
    const res = await fetch(`${API_URL}/api/admin/users/${id}`, {
      method: "PATCH", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ role }), credentials: "include",
    });
    if (res.ok) setUsers((prev) => prev.map((u) => (u.id === id ? { ...u, role } : u)));
  };

  // Auth loading
  if (authLoading) return <FullScreenLoader />;

  // Access denied
  if (!user || user.role !== "admin") return <AccessDenied />;

  return (
    <div className="min-h-screen bg-[#080d19] flex">
      <AdminSidebar view={view} setView={setView} user={user} />
      <div className="flex-1 flex flex-col min-h-screen overflow-hidden">
        <AdminHeader view={view} />
        <main className="flex-1 overflow-y-auto p-8">
          {error && <ErrorBanner message={error} onRetry={fetchData} />}
          {view === "overview" && <OverviewView stats={stats} users={users} credentials={credentials} loading={dataLoading} setView={setView} />}
          {view === "users" && <UsersView users={users} total={userTotal} loading={dataLoading} onRoleChange={updateRole} />}
          {view === "credentials" && <CredentialsView credentials={credentials} total={credTotal} loading={dataLoading} />}
          {view === "proofs" && <CredentialsView credentials={credentials.filter(c => c.cid)} total={credentials.filter(c => c.cid).length} loading={dataLoading} />}
          {view === "settings" && <SettingsView />}
        </main>
      </div>
    </div>
  );
}

// --- Views ---

function OverviewView({ stats, users, credentials, loading, setView }: { stats: Stats | null; users: User[]; credentials: Credential[]; loading: boolean; setView: (v: View) => void }) {
  if (loading) return <SkeletonGrid />;
  const avgScore = credentials.length > 0 ? Math.round(credentials.reduce((a, c) => a + c.score, 0) / credentials.length) : 0;
  return (
    <div className="space-y-8">
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        <StatCard label="Total Users" value={stats?.total_users ?? 0} color="emerald" icon={<svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" /></svg>} />
        <StatCard label="Submissions" value={stats?.total_submissions ?? 0} color="blue" icon={<svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" /></svg>} />
        <StatCard label="Avg Score" value={avgScore} suffix="/100" color="amber" icon={<svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" /></svg>} />
        <StatCard label="On-Chain Proofs" value={credentials.filter(c => c.cid).length} color="purple" icon={<svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M21 7.5l-9-5.25L3 7.5m18 0l-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" /></svg>} />
      </div>

      {/* Recent Users */}
      <Panel title="Recent Users" action={{ label: "View All", onClick: () => setView("users") }}>
        {users.slice(0, 5).map((u) => (
          <div key={u.id} className="flex items-center justify-between py-3 px-1">
            <div className="flex items-center gap-3">
              <img src={u.avatar_url} alt="" className="w-8 h-8 rounded-full" />
              <div>
                <p className="text-[13px] text-white font-medium">{u.username}</p>
                <p className="text-[11px] text-zinc-500">{u.email || "No email"}</p>
              </div>
            </div>
            <span className={`px-2 py-0.5 text-[10px] rounded-full font-medium ${u.role === "admin" ? "bg-emerald-500/10 text-emerald-400" : "bg-white/5 text-zinc-500"}`}>{u.role}</span>
          </div>
        ))}
        {users.length === 0 && <p className="text-zinc-600 text-sm py-4 text-center">No users yet</p>}
      </Panel>

      {/* Recent Submissions */}
      <Panel title="Recent Submissions" action={{ label: "View All", onClick: () => setView("credentials") }}>
        {credentials.slice(0, 5).map((c) => (
          <div key={c.id} className="flex items-center justify-between py-3 px-1">
            <div className="flex items-center gap-3">
              <div className={`w-8 h-8 rounded-lg flex items-center justify-center text-[12px] font-bold ${c.score >= 80 ? "bg-emerald-500/10 text-emerald-400" : c.score >= 50 ? "bg-amber-500/10 text-amber-400" : "bg-red-500/10 text-red-400"}`}>{c.score}</div>
              <div>
                <p className="text-[13px] text-white font-medium truncate max-w-[200px]">{c.name}</p>
                <p className="text-[11px] text-zinc-500">{c.type} &middot; {new Date(c.created_at).toLocaleDateString()}</p>
              </div>
            </div>
            {c.cid && <a href={`https://gateway.lighthouse.storage/ipfs/${c.cid}`} target="_blank" rel="noopener noreferrer" className="text-[11px] text-blue-400 hover:underline">IPFS</a>}
          </div>
        ))}
        {credentials.length === 0 && <p className="text-zinc-600 text-sm py-4 text-center">No submissions yet</p>}
      </Panel>
    </div>
  );
}

function UsersView({ users, total, loading, onRoleChange }: { users: User[]; total: number; loading: boolean; onRoleChange: (id: number, role: string) => void }) {
  if (loading) return <SkeletonTable />;
  return (
    <div className="bg-[#0c1222] border border-white/[0.06] rounded-2xl overflow-hidden">
      <div className="px-6 py-4 border-b border-white/[0.06] flex items-center justify-between">
        <p className="text-white font-semibold text-[14px]">{total} Registered Users</p>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-[13px]">
          <thead><tr className="border-b border-white/[0.06] text-zinc-500">
            <th className="text-left px-6 py-3 font-medium">User</th>
            <th className="text-left px-6 py-3 font-medium">Email</th>
            <th className="text-left px-6 py-3 font-medium">Role</th>
            <th className="text-left px-6 py-3 font-medium">Joined</th>
            <th className="text-left px-6 py-3 font-medium">Action</th>
          </tr></thead>
          <tbody className="divide-y divide-white/[0.04]">
            {users.map((u) => (
              <tr key={u.id} className="hover:bg-white/[0.015] transition-colors">
                <td className="px-6 py-3.5"><div className="flex items-center gap-3"><img src={u.avatar_url} alt="" className="w-7 h-7 rounded-full" /><span className="text-white font-medium">{u.username}</span></div></td>
                <td className="px-6 py-3.5 text-zinc-400">{u.email || "-"}</td>
                <td className="px-6 py-3.5"><span className={`px-2 py-0.5 text-[10px] rounded-full font-medium ${u.role === "admin" ? "bg-emerald-500/10 text-emerald-400" : "bg-white/5 text-zinc-500"}`}>{u.role}</span></td>
                <td className="px-6 py-3.5 text-zinc-500">{new Date(u.created_at).toLocaleDateString()}</td>
                <td className="px-6 py-3.5">
                  {u.role === "user"
                    ? <button onClick={() => onRoleChange(u.id, "admin")} className="text-[11px] text-emerald-400 hover:text-emerald-300">Promote</button>
                    : <button onClick={() => onRoleChange(u.id, "user")} className="text-[11px] text-red-400 hover:text-red-300">Demote</button>}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {users.length === 0 && <p className="text-zinc-600 text-sm py-8 text-center">No users registered yet</p>}
    </div>
  );
}

function CredentialsView({ credentials, total, loading }: { credentials: Credential[]; total: number; loading: boolean }) {
  if (loading) return <SkeletonTable />;
  return (
    <div className="bg-[#0c1222] border border-white/[0.06] rounded-2xl overflow-hidden">
      <div className="px-6 py-4 border-b border-white/[0.06]">
        <p className="text-white font-semibold text-[14px]">{total} Credentials</p>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-[13px]">
          <thead><tr className="border-b border-white/[0.06] text-zinc-500">
            <th className="text-left px-6 py-3 font-medium">ID</th>
            <th className="text-left px-6 py-3 font-medium">Name</th>
            <th className="text-left px-6 py-3 font-medium">Type</th>
            <th className="text-left px-6 py-3 font-medium">Score</th>
            <th className="text-left px-6 py-3 font-medium">CID</th>
            <th className="text-left px-6 py-3 font-medium">Date</th>
          </tr></thead>
          <tbody className="divide-y divide-white/[0.04]">
            {credentials.map((c) => (
              <tr key={c.id} className="hover:bg-white/[0.015] transition-colors">
                <td className="px-6 py-3.5 text-zinc-400 font-mono">#{c.id}</td>
                <td className="px-6 py-3.5 text-white max-w-[200px] truncate">{c.name}</td>
                <td className="px-6 py-3.5"><span className="px-2 py-0.5 text-[10px] rounded-full bg-white/5 text-zinc-400">{c.type}</span></td>
                <td className="px-6 py-3.5"><span className={`font-bold ${c.score >= 80 ? "text-emerald-400" : c.score >= 50 ? "text-amber-400" : "text-red-400"}`}>{c.score}</span></td>
                <td className="px-6 py-3.5">{c.cid ? <a href={`https://gateway.lighthouse.storage/ipfs/${c.cid}`} target="_blank" rel="noopener noreferrer" className="text-blue-400 hover:underline font-mono text-[11px]">{c.cid.slice(0, 14)}...</a> : <span className="text-zinc-700">-</span>}</td>
                <td className="px-6 py-3.5 text-zinc-500">{new Date(c.created_at).toLocaleDateString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {credentials.length === 0 && <p className="text-zinc-600 text-sm py-8 text-center">No credentials yet</p>}
    </div>
  );
}

function SettingsView() {
  return (
    <div className="bg-[#0c1222] border border-white/[0.06] rounded-2xl p-8">
      <h3 className="text-white font-semibold mb-4">System Information</h3>
      <div className="grid grid-cols-2 gap-4 text-[13px]">
        <div><p className="text-zinc-500">Platform</p><p className="text-white mt-1">PoSA v1.0</p></div>
        <div><p className="text-zinc-500">Blockchain</p><p className="text-white mt-1">Flow Mainnet</p></div>
        <div><p className="text-zinc-500">Storage</p><p className="text-white mt-1">Lighthouse + Infura IPFS</p></div>
        <div><p className="text-zinc-500">Database</p><p className="text-white mt-1">Supabase PostgreSQL</p></div>
        <div><p className="text-zinc-500">AI Engine</p><p className="text-white mt-1">Impulse AI + Pattern Analysis</p></div>
        <div><p className="text-zinc-500">Contract</p><p className="text-white mt-1">0xf8a2fcf3389475a1</p></div>
      </div>
    </div>
  );
}

// --- Shared UI ---

function Panel({ title, action, children }: { title: string; action?: { label: string; onClick: () => void }; children: React.ReactNode }) {
  return (
    <div className="bg-[#0c1222] border border-white/[0.06] rounded-2xl">
      <div className="px-6 py-4 border-b border-white/[0.06] flex items-center justify-between">
        <p className="text-white font-semibold text-[14px]">{title}</p>
        {action && <button onClick={action.onClick} className="text-[11px] text-emerald-400 hover:underline">{action.label}</button>}
      </div>
      <div className="px-5 py-2 divide-y divide-white/[0.04]">{children}</div>
    </div>
  );
}

function ErrorBanner({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <div className="mb-6 p-4 bg-red-500/10 border border-red-500/20 rounded-xl flex items-center justify-between">
      <p className="text-red-400 text-[13px]">{message}</p>
      <button onClick={onRetry} className="text-[12px] text-red-300 hover:text-white px-3 py-1 rounded-lg bg-red-500/10 hover:bg-red-500/20 transition-colors">Retry</button>
    </div>
  );
}

function FullScreenLoader() {
  return (
    <div className="min-h-screen bg-[#080d19] flex items-center justify-center">
      <div className="flex flex-col items-center gap-3">
        <div className="w-8 h-8 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
        <p className="text-zinc-500 text-[12px]">Loading dashboard...</p>
      </div>
    </div>
  );
}

function AccessDenied() {
  return (
    <div className="min-h-screen bg-[#080d19] flex items-center justify-center">
      <div className="text-center max-w-sm">
        <div className="w-16 h-16 mx-auto mb-5 rounded-2xl bg-red-500/10 flex items-center justify-center">
          <svg className="w-8 h-8 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" /></svg>
        </div>
        <p className="text-xl font-bold text-white mb-2">Access Denied</p>
        <p className="text-[13px] text-zinc-500 mb-6">You need admin privileges to access this page.</p>
        <a href="/" className="inline-block px-5 py-2 text-[13px] bg-white/5 border border-white/10 text-white rounded-xl hover:bg-white/10 transition-colors">Back to Home</a>
      </div>
    </div>
  );
}

function SkeletonGrid() {
  return (
    <div className="space-y-8">
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        {[1,2,3,4].map(i => <div key={i} className="bg-[#0c1222] border border-white/[0.06] rounded-2xl p-5 h-[130px] animate-pulse" />)}
      </div>
      <div className="bg-[#0c1222] border border-white/[0.06] rounded-2xl h-[300px] animate-pulse" />
    </div>
  );
}

function SkeletonTable() {
  return (
    <div className="bg-[#0c1222] border border-white/[0.06] rounded-2xl overflow-hidden">
      <div className="px-6 py-4 border-b border-white/[0.06]"><div className="h-4 w-32 bg-white/5 rounded animate-pulse" /></div>
      <div className="p-6 space-y-4">
        {[1,2,3,4,5].map(i => <div key={i} className="h-10 bg-white/[0.02] rounded-lg animate-pulse" />)}
      </div>
    </div>
  );
}
