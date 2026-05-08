"use client";

import { useEffect, useState } from "react";
import Navbar from "@/components/Navbar";
import Footer from "@/components/Footer";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

interface Proof {
  id: number;
  user_id: number;
  type: string;
  name: string;
  score: number;
  cid: string;
  tx_hash: string;
  created_at: string;
}

interface ProofsResponse {
  proofs: Proof[] | null;
  total: number;
  limit: number;
  offset: number;
}

function scoreColor(score: number): string {
  if (score >= 80) return "text-green-600 dark:text-green-400";
  if (score >= 50) return "text-yellow-600 dark:text-yellow-400";
  return "text-red-600 dark:text-red-400";
}

function scoreBg(score: number): string {
  if (score >= 80) return "bg-green-100 dark:bg-green-900";
  if (score >= 50) return "bg-yellow-100 dark:bg-yellow-900";
  return "bg-red-100 dark:bg-red-900";
}

export default function ProofsPage() {
  const [proofs, setProofs] = useState<Proof[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(true);
  const limit = 12;

  useEffect(() => {
    setLoading(true);
    fetch(`${API_URL}/api/proofs?limit=${limit}&offset=${offset}`, { credentials: "include" })
      .then((r) => r.json())
      .then((data) => {
        const resp = data.data as ProofsResponse;
        setProofs(resp.proofs || []);
        setTotal(resp.total);
      })
      .catch(() => setProofs([]))
      .finally(() => setLoading(false));
  }, [offset]);

  const totalPages = Math.ceil(total / limit);
  const currentPage = Math.floor(offset / limit) + 1;

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-black flex flex-col">
      <Navbar />

      <main className="flex-1 max-w-6xl mx-auto px-6 py-12 w-full">
        <div className="text-center mb-10">
          <h2 className="text-3xl font-bold text-black dark:text-white mb-3">
            Minted Proofs
          </h2>
          <p className="text-zinc-600 dark:text-zinc-400 max-w-lg mx-auto">
            Verifiable proof-of-skill credentials anchored on the blockchain.
            Each proof links an AI evaluation to an immutable on-chain record.
          </p>
        </div>

        {loading ? (
          <div className="text-center py-20 text-zinc-500">Loading proofs...</div>
        ) : proofs.length === 0 ? (
          <div className="text-center py-20">
            <p className="text-zinc-500 mb-4">No proofs minted yet.</p>
            <a href="/analyze" className="text-blue-600 dark:text-blue-400 hover:underline">
              Submit your work to get started
            </a>
          </div>
        ) : (
          <>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
              {proofs.map((proof) => (
                <div
                  key={proof.id}
                  className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-xl p-6 hover:border-zinc-400 dark:hover:border-zinc-600 transition-colors"
                >
                  {/* Score badge */}
                  <div className="flex items-center justify-between mb-4">
                    <span className={`text-2xl font-bold ${scoreColor(proof.score)}`}>
                      {proof.score}
                    </span>
                    <span className={`px-2.5 py-1 text-xs font-medium rounded-full ${scoreBg(proof.score)} ${scoreColor(proof.score)}`}>
                      {proof.type}
                    </span>
                  </div>

                  {/* Name */}
                  <p className="text-sm font-medium text-zinc-800 dark:text-zinc-200 truncate mb-2">
                    {proof.name}
                  </p>

                  {/* Timestamp */}
                  <p className="text-xs text-zinc-400 mb-4">
                    {new Date(proof.created_at).toLocaleDateString(undefined, {
                      year: "numeric",
                      month: "short",
                      day: "numeric",
                    })}
                  </p>

                  {/* Links */}
                  <div className="flex items-center gap-3 text-xs">
                    {proof.cid && (
                      <a
                        href={`https://gateway.lighthouse.storage/ipfs/${proof.cid}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-600 dark:text-blue-400 hover:underline"
                      >
                        IPFS Report
                      </a>
                    )}
                    {proof.tx_hash && (
                      <a
                        href={`https://flowscan.io/tx/${proof.tx_hash}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-600 dark:text-blue-400 hover:underline"
                      >
                        On-Chain
                      </a>
                    )}
                    <a
                      href={`/verify/${proof.cid || ""}`}
                      className="text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-300"
                    >
                      Verify
                    </a>
                  </div>
                </div>
              ))}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-center gap-4 mt-10">
                <button
                  onClick={() => setOffset(Math.max(0, offset - limit))}
                  disabled={offset === 0}
                  className="px-4 py-2 text-sm border border-zinc-300 dark:border-zinc-700 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:border-zinc-500 transition-colors"
                >
                  Previous
                </button>
                <span className="text-sm text-zinc-500">
                  Page {currentPage} of {totalPages}
                </span>
                <button
                  onClick={() => setOffset(offset + limit)}
                  disabled={offset + limit >= total}
                  className="px-4 py-2 text-sm border border-zinc-300 dark:border-zinc-700 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:border-zinc-500 transition-colors"
                >
                  Next
                </button>
              </div>
            )}
          </>
        )}
      </main>

      <Footer />
    </div>
  );
}
