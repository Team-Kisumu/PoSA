"use client";

import { useEffect, useState } from "react";
import Navbar from "@/components/Navbar";
import Footer from "@/components/Footer";
import { getMe } from "@/lib/api";
import type { UserProfile } from "@/lib/api";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

type Tab = "stats" | "users" | "credentials";

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
  const [tab, setTab] = useState<Tab>("stats");
  const [stats, setStats] = useState<Stats | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [credentials, setCredentials] = useState<Credential[]>([]);

  useEffect(() => {
    getMe().then((u) => {
      setUser(u);
      setLoading(false);
    });
  }, []);

  useEffect(() => {
    if (!user || user.role !== "admin") return;
    if (tab === "stats") {
      fetch(`${API_URL}/api/admin/stats`, { credentials: "include" })
        .then((r) => r.json())
        .then((d) => setStats(d.data));
    } else if (tab === "users") {
      fetch(`${API_URL}/api/admin/users?limit=50`, { credentials: "include" })
        .then((r) => r.json())
        .then((d) => setUsers(d.data?.users || []));
    } else if (tab === "credentials") {
      fetch(`${API_URL}/api/admin/credentials?limit=50`, { credentials: "include" })
        .then((r) => r.json())
        .then((d) => setCredentials(d.data?.credentials || []));
    }
  }, [tab, user]);

  const updateRole = async (id: number, role: string) => {
    await fetch(`${API_URL}/api/admin/users/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ role }),
      credentials: "include",
    });
    setUsers((prev) => prev.map((u) => (u.id === id ? { ...u, role } : u)));
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-zinc-50 dark:bg-black flex flex-col">
        <Navbar />
        <main className="flex-1 flex items-center justify-center">
          <p className="text-zinc-500">Loading...</p>
        </main>
      </div>
    );
  }

  if (!user || user.role !== "admin") {
    return (
      <div className="min-h-screen bg-zinc-50 dark:bg-black flex flex-col">
        <Navbar />
        <main className="flex-1 flex items-center justify-center">
          <div className="text-center">
            <p className="text-2xl font-bold text-red-600 mb-2">Access Denied</p>
            <p className="text-zinc-500">Admin privileges required.</p>
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  const tabClass = (t: Tab) =>
    `px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
      tab === t
        ? "bg-black text-white dark:bg-white dark:text-black"
        : "text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-300"
    }`;

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-black flex flex-col">
      <Navbar />

      <main className="flex-1 max-w-6xl mx-auto px-6 py-12 w-full">
        <div className="flex items-center justify-between mb-8">
          <h2 className="text-2xl font-bold text-black dark:text-white">Admin Dashboard</h2>
          <div className="flex gap-2">
            <button className={tabClass("stats")} onClick={() => setTab("stats")}>Stats</button>
            <button className={tabClass("users")} onClick={() => setTab("users")}>Users</button>
            <button className={tabClass("credentials")} onClick={() => setTab("credentials")}>Credentials</button>
          </div>
        </div>

        {/* Stats Tab */}
        {tab === "stats" && stats && (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-xl p-8 text-center">
              <p className="text-4xl font-bold text-black dark:text-white">{stats.total_users}</p>
              <p className="text-sm text-zinc-500 mt-2">Registered Users</p>
            </div>
            <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-xl p-8 text-center">
              <p className="text-4xl font-bold text-black dark:text-white">{stats.total_submissions}</p>
              <p className="text-sm text-zinc-500 mt-2">Total Submissions</p>
            </div>
          </div>
        )}

        {/* Users Tab */}
        {tab === "users" && (
          <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-zinc-50 dark:bg-zinc-800 border-b border-zinc-200 dark:border-zinc-700">
                <tr>
                  <th className="text-left px-6 py-3 font-medium text-zinc-500">User</th>
                  <th className="text-left px-6 py-3 font-medium text-zinc-500">Email</th>
                  <th className="text-left px-6 py-3 font-medium text-zinc-500">Role</th>
                  <th className="text-left px-6 py-3 font-medium text-zinc-500">Joined</th>
                  <th className="text-left px-6 py-3 font-medium text-zinc-500">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-100 dark:divide-zinc-800">
                {users.map((u) => (
                  <tr key={u.id}>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        {u.avatar_url && <img src={u.avatar_url} alt="" className="w-7 h-7 rounded-full" />}
                        <span className="font-medium text-zinc-800 dark:text-zinc-200">{u.username}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-zinc-500">{u.email || "-"}</td>
                    <td className="px-6 py-4">
                      <span className={`px-2 py-0.5 text-xs rounded-full ${
                        u.role === "admin"
                          ? "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200"
                          : "bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400"
                      }`}>
                        {u.role}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-zinc-500">
                      {new Date(u.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4">
                      {u.role === "user" ? (
                        <button
                          onClick={() => updateRole(u.id, "admin")}
                          className="text-xs text-blue-600 hover:underline"
                        >
                          Make Admin
                        </button>
                      ) : (
                        <button
                          onClick={() => updateRole(u.id, "user")}
                          className="text-xs text-red-600 hover:underline"
                        >
                          Remove Admin
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {users.length === 0 && (
              <p className="text-center py-8 text-zinc-500">No users registered yet.</p>
            )}
          </div>
        )}

        {/* Credentials Tab */}
        {tab === "credentials" && (
          <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-zinc-50 dark:bg-zinc-800 border-b border-zinc-200 dark:border-zinc-700">
                <tr>
                  <th className="text-left px-6 py-3 font-medium text-zinc-500">ID</th>
                  <th className="text-left px-6 py-3 font-medium text-zinc-500">Name</th>
                  <th className="text-left px-6 py-3 font-medium text-zinc-500">Type</th>
                  <th className="text-left px-6 py-3 font-medium text-zinc-500">Score</th>
                  <th className="text-left px-6 py-3 font-medium text-zinc-500">CID</th>
                  <th className="text-left px-6 py-3 font-medium text-zinc-500">Date</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-100 dark:divide-zinc-800">
                {credentials.map((c) => (
                  <tr key={c.id}>
                    <td className="px-6 py-4 text-zinc-800 dark:text-zinc-200">#{c.id}</td>
                    <td className="px-6 py-4 text-zinc-800 dark:text-zinc-200 max-w-48 truncate">{c.name}</td>
                    <td className="px-6 py-4">
                      <span className="px-2 py-0.5 text-xs rounded-full bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400">
                        {c.type}
                      </span>
                    </td>
                    <td className="px-6 py-4 font-medium">
                      <span className={c.score >= 80 ? "text-green-600" : c.score >= 50 ? "text-yellow-600" : "text-red-600"}>
                        {c.score}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      {c.cid ? (
                        <a href={`https://gateway.lighthouse.storage/ipfs/${c.cid}`} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline text-xs font-mono">
                          {c.cid.slice(0, 12)}...
                        </a>
                      ) : (
                        <span className="text-zinc-400">-</span>
                      )}
                    </td>
                    <td className="px-6 py-4 text-zinc-500">{new Date(c.created_at).toLocaleDateString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            {credentials.length === 0 && (
              <p className="text-center py-8 text-zinc-500">No credentials minted yet.</p>
            )}
          </div>
        )}
      </main>

      <Footer />
    </div>
  );
}
