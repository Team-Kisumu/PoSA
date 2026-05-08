"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Navbar from "@/components/Navbar";
import Footer from "@/components/Footer";
import { verifyCID } from "@/lib/api";
import type { VerifyResult } from "@/lib/api";

export default function VerifyCIDPage() {
  const params = useParams();
  const cid = params.cid as string;
  const [loading, setLoading] = useState(true);
  const [result, setResult] = useState<VerifyResult | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!cid) return;
    setLoading(true);
    verifyCID(cid)
      .then(setResult)
      .catch((e) => setError(e instanceof Error ? e.message : "Verification failed"))
      .finally(() => setLoading(false));
  }, [cid]);

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-black flex flex-col">
      <Navbar />
      <main className="flex-1 max-w-2xl mx-auto px-6 py-12 w-full">
        <div className="text-center mb-10">
          <h2 className="text-3xl font-bold text-black dark:text-white mb-3">Credential Verification</h2>
          <p className="text-zinc-600 dark:text-zinc-400 text-sm font-mono break-all">{cid}</p>
        </div>

        {loading && (
          <div className="text-center py-12">
            <div className="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto mb-3" />
            <p className="text-zinc-500 text-sm">Verifying credential...</p>
          </div>
        )}

        {error && (
          <div className="p-4 bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg text-sm text-red-700 dark:text-red-300">{error}</div>
        )}

        {result && !loading && (
          <div className="space-y-4">
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
                  <p className={`text-lg font-bold ${result.valid ? "text-green-800 dark:text-green-200" : "text-red-800 dark:text-red-200"}`}>
                    {result.valid ? "Credential Verified" : "Credential Not Found"}
                  </p>
                  <p className={`text-sm ${result.valid ? "text-green-600 dark:text-green-400" : "text-red-600 dark:text-red-400"}`}>
                    {result.valid ? "This CID has a valid on-chain proof record" : "No credential exists on-chain for this CID"}
                  </p>
                </div>
              </div>
            </div>

            {result.valid && result.credential && (
              <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-700 rounded-lg">
                <div className="px-6 py-4 border-b border-zinc-200 dark:border-zinc-700">
                  <h3 className="font-semibold text-black dark:text-white">On-Chain Record</h3>
                </div>
                <div className="px-6 py-4 space-y-3 text-sm">
                  <div className="flex justify-between"><span className="text-zinc-500">ID</span><span className="font-mono text-zinc-800 dark:text-zinc-200">#{result.credential.id}</span></div>
                  <div className="flex justify-between"><span className="text-zinc-500">Score</span><span className={`font-bold ${result.credential.score >= 80 ? "text-green-600" : result.credential.score >= 50 ? "text-yellow-600" : "text-red-600"}`}>{result.credential.score}/100</span></div>
                  <div className="flex justify-between"><span className="text-zinc-500">Submitter</span><span className="font-mono text-zinc-800 dark:text-zinc-200">{result.credential.submitter}</span></div>
                  <div className="flex justify-between"><span className="text-zinc-500">IPFS</span><a href={`https://gateway.lighthouse.storage/ipfs/${cid}`} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">View Report</a></div>
                </div>
              </div>
            )}

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

        <div className="mt-8 text-center">
          <a href="/verify" className="text-sm text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-300">Verify another credential</a>
        </div>
      </main>
      <Footer />
    </div>
  );
}
