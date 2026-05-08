"use client";

import { useEffect, useState } from "react";
import { QRCodeSVG } from "qrcode.react";
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

function scoreColor(score: number) {
  if (score >= 80) return { text: "text-emerald-400", bg: "bg-emerald-500/20", border: "border-emerald-500/30" };
  if (score >= 50) return { text: "text-amber-400", bg: "bg-amber-500/20", border: "border-amber-500/30" };
  return { text: "text-red-400", bg: "bg-red-500/20", border: "border-red-500/30" };
}

function getVerifyUrl(proof: Proof): string {
  if (proof.cid) return `${typeof window !== "undefined" ? window.location.origin : ""}/verify/${proof.cid}`;
  return `${typeof window !== "undefined" ? window.location.origin : ""}/proofs`;
}

export default function ProofsPage() {
  const [proofs, setProofs] = useState<Proof[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(true);
  const [selected, setSelected] = useState<Proof | null>(null);
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
    <div className="min-h-screen bg-zinc-50 dark:bg-[#080d19] flex flex-col">
      <Navbar />

      <main className="flex-1 max-w-6xl mx-auto px-4 sm:px-6 py-12 w-full">
        <div className="text-center mb-10">
          <h2 className="text-3xl font-bold text-black dark:text-white mb-3">Minted Proofs</h2>
          <p className="text-zinc-600 dark:text-zinc-400 max-w-lg mx-auto">
            Verifiable proof-of-skill credentials. Scan the QR code to verify any credential independently.
          </p>
        </div>

        {loading ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {[1, 2, 3, 4, 5, 6].map((i) => (
              <div key={i} className="h-[320px] rounded-2xl bg-white/5 dark:bg-white/[0.03] border border-zinc-200 dark:border-white/[0.06] animate-pulse" />
            ))}
          </div>
        ) : proofs.length === 0 ? (
          <div className="text-center py-20">
            <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-zinc-100 dark:bg-white/5 flex items-center justify-center">
              <svg className="w-8 h-8 text-zinc-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" /></svg>
            </div>
            <p className="text-zinc-500 mb-4">No proofs minted yet.</p>
            <a href="/analyze" className="text-emerald-500 hover:underline text-sm">Submit your work to get started</a>
          </div>
        ) : (
          <>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
              {proofs.map((proof) => (
                <ProofCard key={proof.id} proof={proof} onClick={() => setSelected(proof)} />
              ))}
            </div>

            {totalPages > 1 && (
              <div className="flex items-center justify-center gap-4 mt-10">
                <button onClick={() => setOffset(Math.max(0, offset - limit))} disabled={offset === 0} className="px-4 py-2 text-sm border border-zinc-300 dark:border-zinc-700 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:border-zinc-500 transition-colors text-zinc-700 dark:text-zinc-300">Previous</button>
                <span className="text-sm text-zinc-500">Page {currentPage} of {totalPages}</span>
                <button onClick={() => setOffset(offset + limit)} disabled={offset + limit >= total} className="px-4 py-2 text-sm border border-zinc-300 dark:border-zinc-700 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:border-zinc-500 transition-colors text-zinc-700 dark:text-zinc-300">Next</button>
              </div>
            )}
          </>
        )}
      </main>

      <Footer />

      {/* Detail Overlay */}
      {selected && <ProofOverlay proof={selected} onClose={() => setSelected(null)} />}
    </div>
  );
}

// --- Glassmorphic Proof Card ---
function ProofCard({ proof, onClick }: { proof: Proof; onClick: () => void }) {
  const sc = scoreColor(proof.score);
  const verifyUrl = getVerifyUrl(proof);

  return (
    <button
      onClick={onClick}
      className="group relative w-full text-left rounded-2xl overflow-hidden border border-zinc-200 dark:border-white/[0.08] bg-white dark:bg-white/[0.03] backdrop-blur-xl hover:border-zinc-400 dark:hover:border-white/[0.15] transition-all duration-300 hover:shadow-lg hover:shadow-black/5 dark:hover:shadow-emerald-500/5"
    >
      {/* Top section: QR + Score */}
      <div className="p-5 pb-4 flex items-start justify-between">
        <div className="w-20 h-20 rounded-xl bg-white p-1.5 shadow-sm">
          <QRCodeSVG value={verifyUrl} size={68} level="M" bgColor="white" fgColor="#18181b" />
        </div>
        <div className={`px-3 py-1.5 rounded-xl ${sc.bg} border ${sc.border}`}>
          <span className={`text-xl font-bold ${sc.text}`}>{proof.score}</span>
        </div>
      </div>

      {/* Info */}
      <div className="px-5 pb-5">
        <p className="text-[13px] font-semibold text-zinc-800 dark:text-white truncate mb-1">{proof.name}</p>
        <div className="flex items-center gap-2 mb-3">
          <span className="px-2 py-0.5 text-[10px] rounded-full bg-zinc-100 dark:bg-white/5 text-zinc-500 dark:text-zinc-400 font-medium">{proof.type}</span>
          <span className="text-[11px] text-zinc-400">{new Date(proof.created_at).toLocaleDateString()}</span>
        </div>

        {/* Bottom links */}
        <div className="flex items-center gap-3 text-[11px]">
          {proof.cid && (
            <span className="flex items-center gap-1 text-blue-500">
              <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101" /><path strokeLinecap="round" strokeLinejoin="round" d="M10.172 13.828a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" /></svg>
              IPFS
            </span>
          )}
          {proof.tx_hash && (
            <span className="flex items-center gap-1 text-purple-500">
              <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M21 7.5l-9-5.25L3 7.5m18 0l-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" /></svg>
              On-Chain
            </span>
          )}
        </div>
      </div>

      {/* Hover hint */}
      <div className="absolute bottom-0 left-0 right-0 h-8 bg-gradient-to-t from-zinc-100/80 dark:from-white/[0.02] to-transparent flex items-end justify-center pb-1.5 opacity-0 group-hover:opacity-100 transition-opacity">
        <span className="text-[10px] text-zinc-400">Tap to view details</span>
      </div>
    </button>
  );
}

// --- Full-screen Detail Overlay ---
function ProofOverlay({ proof, onClose }: { proof: Proof; onClose: () => void }) {
  const sc = scoreColor(proof.score);
  const verifyUrl = getVerifyUrl(proof);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm" onClick={onClose}>
      <div className="relative w-full max-w-lg bg-white dark:bg-[#0f1425] rounded-3xl shadow-2xl border border-zinc-200 dark:border-white/[0.08] overflow-hidden" onClick={(e) => e.stopPropagation()}>
        {/* Close button */}
        <button onClick={onClose} className="absolute top-4 right-4 z-10 w-8 h-8 rounded-full bg-zinc-100 dark:bg-white/10 flex items-center justify-center text-zinc-500 hover:text-zinc-800 dark:hover:text-white transition-colors">
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>

        {/* Header with QR */}
        <div className="p-8 pb-6 flex flex-col items-center text-center border-b border-zinc-100 dark:border-white/[0.06]">
          <div className="w-40 h-40 rounded-2xl bg-white p-3 shadow-lg mb-5">
            <QRCodeSVG value={verifyUrl} size={152} level="H" bgColor="white" fgColor="#18181b" />
          </div>
          <div className={`px-4 py-2 rounded-xl ${sc.bg} border ${sc.border} mb-3`}>
            <span className={`text-3xl font-bold ${sc.text}`}>{proof.score}</span>
            <span className={`text-sm ${sc.text} ml-1`}>/100</span>
          </div>
          <p className="text-[11px] text-zinc-400">Scan QR to verify this credential</p>
        </div>

        {/* Details */}
        <div className="p-6 space-y-4">
          <DetailRow label="Submission" value={proof.name} mono />
          <DetailRow label="Type" value={proof.type} />
          <DetailRow label="Date" value={new Date(proof.created_at).toLocaleString()} />
          {proof.cid && (
            <DetailRow label="IPFS CID" value={proof.cid} mono link={`https://gateway.lighthouse.storage/ipfs/${proof.cid}`} />
          )}
          {proof.tx_hash && (
            <DetailRow label="Transaction" value={proof.tx_hash} mono link={`https://flowscan.io/tx/${proof.tx_hash}`} />
          )}
          <DetailRow label="Proof ID" value={`#${proof.id}`} />
        </div>

        {/* Footer */}
        <div className="px-6 pb-6 flex gap-3">
          {proof.cid && (
            <a href={`https://gateway.lighthouse.storage/ipfs/${proof.cid}`} target="_blank" rel="noopener noreferrer" className="flex-1 py-2.5 text-center text-[13px] font-medium bg-blue-500/10 text-blue-500 rounded-xl hover:bg-blue-500/20 transition-colors">
              View Report
            </a>
          )}
          <a href={verifyUrl} className="flex-1 py-2.5 text-center text-[13px] font-medium bg-emerald-500/10 text-emerald-500 rounded-xl hover:bg-emerald-500/20 transition-colors">
            Verify
          </a>
        </div>
      </div>
    </div>
  );
}

function DetailRow({ label, value, mono, link }: { label: string; value: string; mono?: boolean; link?: string }) {
  return (
    <div className="flex items-start justify-between gap-4">
      <span className="text-[12px] text-zinc-500 dark:text-zinc-400 shrink-0">{label}</span>
      {link ? (
        <a href={link} target="_blank" rel="noopener noreferrer" className={`text-[12px] text-blue-500 hover:underline text-right truncate max-w-[250px] ${mono ? "font-mono" : ""}`}>{value}</a>
      ) : (
        <span className={`text-[12px] text-zinc-800 dark:text-zinc-200 text-right truncate max-w-[250px] ${mono ? "font-mono" : ""}`}>{value}</span>
      )}
    </div>
  );
}
