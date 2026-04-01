"use client";

import { useState } from "react";
import type { EvaluationResult, MintResult } from "@/lib/api";

interface Props {
  result: EvaluationResult;
  filename: string;
  onReset: () => void;
}

// Score color based on value.
function scoreColor(score: number): string {
  if (score >= 80) return "text-green-600 dark:text-green-400";
  if (score >= 50) return "text-yellow-600 dark:text-yellow-400";
  return "text-red-600 dark:text-red-400";
}

function scoreBg(score: number): string {
  if (score >= 80) return "bg-green-500";
  if (score >= 50) return "bg-yellow-500";
  return "bg-red-500";
}

// Severity badge colors.
function severityBadge(type: string): string {
  switch (type) {
    case "security":
      return "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200";
    case "quality":
      return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200";
    case "logic":
      return "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200";
    case "style":
      return "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200";
    case "formatting":
      return "bg-zinc-100 text-zinc-800 dark:bg-zinc-800 dark:text-zinc-200";
    case "deprecated":
      return "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200";
    case "incompleteness":
      return "bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200";
    case "obsolete":
      return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200";
    default:
      return "bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300";
  }
}

export default function ResultsView({ result, filename, onReset }: Props) {
  const [minting, setMinting] = useState(false);
  const [mintResult, setMintResult] = useState<MintResult | null>(null);
  const [mintError, setMintError] = useState("");

  const handleMint = async () => {
    setMinting(true);
    setMintError("");
    try {
      // In the full pipeline, this calls the backend /api/mint endpoint.
      // For now, simulate a successful mint with placeholder data.
      const simulated: MintResult = {
        cid: "Qm" + Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15) + "abcdefghijklmnop",
        tx_hash: "0x" + Math.random().toString(16).substring(2, 18) + "..." + Math.random().toString(16).substring(2, 10),
        score: result.score,
        submitter: "0xf8a2fcf3389475a1",
      };
      // Simulate network delay.
      await new Promise((r) => setTimeout(r, 1500));
      setMintResult(simulated);
    } catch (err) {
      setMintError(err instanceof Error ? err.message : "Minting failed");
    } finally {
      setMinting(false);
    }
  };

  const hasDetails = result.issues.length > 0 || result.suggestions.length > 0;

  return (
    <div className="w-full max-w-6xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-black dark:text-white">Evaluation Results</h3>
          <p className="text-sm text-zinc-500">{filename}</p>
        </div>
        <button
          onClick={onReset}
          className="text-sm text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-300"
        >
          New Submission
        </button>
      </div>

      {/* Two-column layout: score + files on left, issues + suggestions on right */}
      <div className={`grid gap-6 ${hasDetails ? "grid-cols-1 lg:grid-cols-2" : "grid-cols-1 max-w-2xl"}`}>
        {/* Left column: Score + File breakdown */}
        <div className="space-y-6">
          {/* Score display */}
          <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-700 rounded-lg p-6">
            <div className="flex items-center gap-6">
              <div className="relative w-24 h-24 shrink-0">
                <svg className="w-24 h-24 -rotate-90" viewBox="0 0 100 100">
                  <circle cx="50" cy="50" r="42" fill="none" stroke="currentColor" strokeWidth="8"
                    className="text-zinc-200 dark:text-zinc-700" />
                  <circle cx="50" cy="50" r="42" fill="none" strokeWidth="8"
                    strokeDasharray={`${result.score * 2.64} 264`}
                    strokeLinecap="round"
                    className={scoreBg(result.score).replace("bg-", "text-")} />
                </svg>
                <span className={`absolute inset-0 flex items-center justify-center text-2xl font-bold ${scoreColor(result.score)}`}>
                  {result.score}
                </span>
              </div>
              <div>
                <p className={`text-3xl font-bold ${scoreColor(result.score)}`}>
                  {result.score}/100
                </p>
                <p className="text-sm text-zinc-500 mt-1">
                  {result.issues.length === 0
                    ? "No issues found"
                    : `${result.issues.length} issue${result.issues.length > 1 ? "s" : ""} found`}
                </p>
              </div>
            </div>
          </div>

          {/* Per-file breakdown (repo analysis) */}
          {result.file_scores && result.file_scores.length > 0 && (
            <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-700 rounded-lg">
              <div className="px-6 py-4 border-b border-zinc-200 dark:border-zinc-700">
                <h4 className="font-semibold text-black dark:text-white">
                  Files Analyzed ({result.files_analyzed})
                </h4>
              </div>
              <ul className="divide-y divide-zinc-100 dark:divide-zinc-800 max-h-96 overflow-y-auto">
                {result.file_scores.map((fs, i) => (
                  <li key={i} className="px-6 py-3 flex items-center justify-between">
                    <div className="flex-1 min-w-0">
                      <p className="text-sm text-zinc-800 dark:text-zinc-200 font-mono truncate">{fs.path}</p>
                      {fs.language && (
                        <p className="text-xs text-zinc-400 mt-0.5">{fs.language}</p>
                      )}
                    </div>
                    <div className="flex items-center gap-3 ml-4">
                      {fs.issues > 0 && (
                        <span className="text-xs text-zinc-500">
                          {fs.issues} issue{fs.issues > 1 ? "s" : ""}
                        </span>
                      )}
                      <span className={`text-sm font-medium ${scoreColor(fs.score)}`}>
                        {fs.score}
                      </span>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>

        {/* Right column: Issues + Suggestions */}
        {hasDetails && (
          <div className="space-y-6">
            {/* Issues list */}
            {result.issues.length > 0 && (
              <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-700 rounded-lg">
                <div className="px-6 py-4 border-b border-zinc-200 dark:border-zinc-700">
                  <h4 className="font-semibold text-black dark:text-white">Issues</h4>
                </div>
                <ul className="divide-y divide-zinc-100 dark:divide-zinc-800 max-h-96 overflow-y-auto">
                  {result.issues.map((issue, i) => (
                    <li key={i} className="px-6 py-3 flex items-start gap-3">
                      <span className={`inline-block px-2 py-0.5 text-xs font-medium rounded-full mt-0.5 shrink-0 ${severityBadge(issue.issue_type)}`}>
                        {issue.issue_type}
                      </span>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm text-zinc-800 dark:text-zinc-200">{issue.message}</p>
                        {issue.line && (
                          <p className="text-xs text-zinc-400 mt-0.5">Line {issue.line}</p>
                        )}
                      </div>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {/* Suggestions */}
            {result.suggestions.length > 0 && (
              <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-700 rounded-lg">
                <div className="px-6 py-4 border-b border-zinc-200 dark:border-zinc-700">
                  <h4 className="font-semibold text-black dark:text-white">Suggestions</h4>
                </div>
                <ul className="px-6 py-4 space-y-2">
                  {result.suggestions.map((s, i) => (
                    <li key={i} className="flex items-start gap-2 text-sm text-zinc-600 dark:text-zinc-400">
                      <span className="text-blue-500 mt-0.5 shrink-0">{"\u2192"}</span>
                      <span>{s}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Mint Proof section (full width) */}
      <div className="max-w-2xl mx-auto">
        {!mintResult ? (
          <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-700 rounded-lg p-6">
            <h4 className="font-semibold text-black dark:text-white mb-2">Anchor on Blockchain</h4>
            <p className="text-sm text-zinc-500 mb-4">
              Store the evaluation report on IPFS/Filecoin and mint a verifiable credential on Flow.
            </p>
            {mintError && (
              <div className="mb-4 p-3 bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg text-sm text-red-700 dark:text-red-300">
                {mintError}
              </div>
            )}
            <button
              onClick={handleMint}
              disabled={minting}
              className="w-full py-3 px-6 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {minting ? "Minting..." : "Mint Proof"}
            </button>
          </div>
        ) : (
          <div className="bg-green-50 dark:bg-green-950 border border-green-200 dark:border-green-800 rounded-lg p-6 space-y-3">
            <h4 className="font-semibold text-green-800 dark:text-green-200">Proof Minted</h4>
            <div className="space-y-2 text-sm">
              <div>
                <span className="text-green-600 dark:text-green-400 font-medium">IPFS CID: </span>
                <a
                  href={`https://gateway.lighthouse.storage/ipfs/${mintResult.cid}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 dark:text-blue-400 hover:underline break-all"
                >
                  {mintResult.cid}
                </a>
              </div>
              <div>
                <span className="text-green-600 dark:text-green-400 font-medium">Transaction: </span>
                <a
                  href={`https://testnet.flowscan.io/tx/${mintResult.tx_hash}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 dark:text-blue-400 hover:underline break-all"
                >
                  {mintResult.tx_hash}
                </a>
              </div>
              <div>
                <span className="text-green-600 dark:text-green-400 font-medium">Score: </span>
                <span className="text-green-800 dark:text-green-200">{mintResult.score}/100</span>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
