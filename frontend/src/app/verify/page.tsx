"use client";

import { useState } from "react";
import Link from "next/link";
import { verifyCID } from "@/lib/api";
import type { VerifyResult } from "@/lib/api";

export default function VerifyPage() {
  const [cid, setCid] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState<VerifyResult | null>(null);

  const handleVerify = async () => {
    const trimmed = cid.trim();
    if (!trimmed) {
      setError("Please enter a CID");
      return;
    }

    setError("");
    setResult(null);
    setLoading(true);

    try {
      const res = await verifyCID(trimmed);
      setResult(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Verification failed");
    } finally {
      setLoading(false);
    }
  };

  const formatTimestamp = (ts: number): string => {
    if (!ts) return "Unknown";
    const date = new Date(ts * 1000);
    return date.toLocaleString();
  };

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-black">
      <header className="border-b border-zinc-200 dark:border-zinc-800">
        <div className="max-w-4xl mx-auto px-6 py-4 flex items-center justify-between">
          <Link href="/" className="group">
            <h1 className="text-xl font-bold text-black dark:text-white group-hover:text-zinc-600">PoSA</h1>
            <p className="text-xs text-zinc-500">Proof-of-Skill AI</p>
          </Link>
          <span className="text-xs text-zinc-400">Credential Verification</span>
        </div>
      </header>

      <main className="max-w-2xl mx-auto px-6 py-12">
        <div className="text-center mb-10">
          <h2 className="text-3xl font-bold text-black dark:text-white mb-3">
            Verify a Credential
          </h2>
          <p className="text-zinc-600 dark:text-zinc-400 max-w-lg mx-auto">
            Enter an IPFS CID to verify the credential against the on-chain record
            and retrieve the evaluation report.
          </p>
        </div>

        {/* CID input */}
        <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-700 rounded-lg p-6">
          <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-2">
            Content Identifier (CID)
          </label>
          <div className="flex gap-3">
            <input
              type="text"
              value={cid}
              onChange={(e) => { setCid(e.target.value); setError(""); setResult(null); }}
              onKeyDown={(e) => e.key === "Enter" && handleVerify()}
              placeholder="QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco"
              className="flex-1 p-3 border border-zinc-300 dark:border-zinc-600 rounded-lg bg-zinc-50 dark:bg-zinc-800 text-zinc-800 dark:text-zinc-200 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              onClick={handleVerify}
              disabled={loading}
              className="px-6 py-3 bg-black text-white rounded-lg font-medium hover:bg-zinc-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors dark:bg-white dark:text-black dark:hover:bg-zinc-200"
            >
              {loading ? "Verifying..." : "Verify"}
            </button>
          </div>

          {error && (
            <div className="mt-4 p-3 bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg text-sm text-red-700 dark:text-red-300">
              {error}
            </div>
          )}
        </div>

        {/* Results */}
        {result && (
          <div className="mt-6 space-y-4">
            {/* Validity status */}
            <div className={`border rounded-lg p-6 ${
              result.valid
                ? "bg-green-50 dark:bg-green-950 border-green-200 dark:border-green-800"
                : "bg-red-50 dark:bg-red-950 border-red-200 dark:border-red-800"
            }`}>
              <div className="flex items-center gap-3">
                <span className={`text-3xl ${result.valid ? "text-green-600" : "text-red-600"}`}>
                  {result.valid ? "\u2713" : "\u2717"}
                </span>
                <div>
                  <p className={`text-lg font-bold ${
                    result.valid
                      ? "text-green-800 dark:text-green-200"
                      : "text-red-800 dark:text-red-200"
                  }`}>
                    {result.valid ? "Credential Verified" : "Credential Not Found"}
                  </p>
                  <p className={`text-sm ${
                    result.valid
                      ? "text-green-600 dark:text-green-400"
                      : "text-red-600 dark:text-red-400"
                  }`}>
                    {result.valid
                      ? "This CID has a valid on-chain proof record"
                      : "No credential exists on-chain for this CID"}
                  </p>
                </div>
              </div>
            </div>

            {/* Credential details */}
            {result.valid && result.credential && (
              <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-700 rounded-lg">
                <div className="px-6 py-4 border-b border-zinc-200 dark:border-zinc-700">
                  <h3 className="font-semibold text-black dark:text-white">On-Chain Record</h3>
                </div>
                <div className="px-6 py-4 space-y-3 text-sm">
                  <div className="flex justify-between">
                    <span className="text-zinc-500">Credential ID</span>
                    <span className="font-mono text-zinc-800 dark:text-zinc-200">#{result.credential.id}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-zinc-500">Score</span>
                    <span className={`font-bold ${
                      result.credential.score >= 80 ? "text-green-600" :
                      result.credential.score >= 50 ? "text-yellow-600" : "text-red-600"
                    }`}>
                      {result.credential.score}/100
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-zinc-500">Submitter</span>
                    <a
                      href={`https://testnet.flowscan.io/account/${result.credential.submitter}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="font-mono text-blue-600 dark:text-blue-400 hover:underline"
                    >
                      {result.credential.submitter}
                    </a>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-zinc-500">Timestamp</span>
                    <span className="text-zinc-800 dark:text-zinc-200">
                      {formatTimestamp(result.credential.timestamp)}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-zinc-500">IPFS Report</span>
                    <a
                      href={`https://gateway.lighthouse.storage/ipfs/${result.credential.cid}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-600 dark:text-blue-400 hover:underline"
                    >
                      View on IPFS
                    </a>
                  </div>
                </div>
              </div>
            )}

            {/* Report preview */}
            {result.report && (
              <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-700 rounded-lg">
                <div className="px-6 py-4 border-b border-zinc-200 dark:border-zinc-700">
                  <h3 className="font-semibold text-black dark:text-white">Evaluation Report</h3>
                </div>
                <pre className="px-6 py-4 text-xs font-mono text-zinc-600 dark:text-zinc-400 overflow-x-auto max-h-64 overflow-y-auto">
                  {JSON.stringify(result.report, null, 2)}
                </pre>
              </div>
            )}
          </div>
        )}
      </main>

      <footer className="border-t border-zinc-200 dark:border-zinc-800 mt-12">
        <div className="max-w-4xl mx-auto px-6 py-4 text-center text-xs text-zinc-400">
          PoSA {"\u00A9"} {new Date().getFullYear()} {"\u2014"} Impulse AI {"\u00B7"} Filecoin {"\u00B7"} Flow
        </div>
      </footer>
    </div>
  );
}
