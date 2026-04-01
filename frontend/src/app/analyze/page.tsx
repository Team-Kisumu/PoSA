"use client";

import { useState } from "react";
import Navbar from "@/components/Navbar";
import Footer from "@/components/Footer";
import UploadForm from "@/components/UploadForm";
import ResultsView from "@/components/ResultsView";
import type { EvaluationResult } from "@/lib/api";

interface SubmissionResult {
  evaluation: EvaluationResult;
  filename: string;
}

export default function AnalyzePage() {
  const [submission, setSubmission] = useState<SubmissionResult | null>(null);

  const handleResult = (evaluation: EvaluationResult, filename: string) => {
    setSubmission({ evaluation, filename });
  };

  const handleReset = () => setSubmission(null);

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-black flex flex-col">
      <Navbar />

      <main className="flex-1 max-w-6xl mx-auto px-6 py-12 w-full">
        {!submission ? (
          <>
            <div className="text-center mb-10">
              <h2 className="text-3xl font-bold text-black dark:text-white mb-3">
                Submit Your Work
              </h2>
              <p className="text-zinc-600 dark:text-zinc-400 max-w-lg mx-auto">
                Upload a file, paste code, or enter a GitHub repo link.
                Our AI evaluates your work and issues a verifiable credential on the blockchain.
              </p>
            </div>

            <UploadForm onResult={handleResult} />

            <div className="mt-16 grid grid-cols-1 md:grid-cols-3 gap-6 text-center max-w-4xl mx-auto">
              <div className="p-6 rounded-lg border border-zinc-200 dark:border-zinc-800">
                <h3 className="font-semibold text-black dark:text-white mb-2">AI Analysis</h3>
                <p className="text-sm text-zinc-500">
                  Pattern-based + Impulse AI evaluation across 25 languages
                </p>
              </div>
              <div className="p-6 rounded-lg border border-zinc-200 dark:border-zinc-800">
                <h3 className="font-semibold text-black dark:text-white mb-2">Decentralized Storage</h3>
                <p className="text-sm text-zinc-500">
                  Reports stored on IPFS/Filecoin via Lighthouse with CID verification
                </p>
              </div>
              <div className="p-6 rounded-lg border border-zinc-200 dark:border-zinc-800">
                <h3 className="font-semibold text-black dark:text-white mb-2">Blockchain Proof</h3>
                <p className="text-sm text-zinc-500">
                  CID anchored on Flow with tamper-proof credential minting
                </p>
              </div>
            </div>
          </>
        ) : (
          <ResultsView
            result={submission.evaluation}
            filename={submission.filename}
            onReset={handleReset}
          />
        )}
      </main>

      <Footer />
    </div>
  );
}
