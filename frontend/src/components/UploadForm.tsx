"use client";

import { useState, useRef, useCallback } from "react";
import { submitFile, submitRepo } from "@/lib/api";

type Tab = "upload" | "paste" | "repo";

interface Result {
  success: boolean;
  data?: Record<string, unknown>;
  error?: { code: string; message: string };
}

export default function UploadForm() {
  const [tab, setTab] = useState<Tab>("upload");
  const [file, setFile] = useState<File | null>(null);
  const [code, setCode] = useState("");
  const [repoUrl, setRepoUrl] = useState("");
  const [dragging, setDragging] = useState(false);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<Result | null>(null);
  const [error, setError] = useState("");
  const fileRef = useRef<HTMLInputElement>(null);

  // --- Drag and drop handlers ---
  const onDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragging(true);
  }, []);

  const onDragLeave = useCallback(() => setDragging(false), []);

  const onDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragging(false);
    const dropped = e.dataTransfer.files[0];
    if (dropped) setFile(dropped);
  }, []);

  // --- Validation ---
  const validateRepo = (url: string): string | null => {
    if (!url.trim()) return "Repository URL is required";
    if (!url.startsWith("https://github.com/")) return "Must be a GitHub HTTPS URL";
    const parts = url.replace("https://github.com/", "").split("/");
    if (parts.length < 2 || !parts[0] || !parts[1]) return "URL must include owner/repo";
    return null;
  };

  // --- Submit handlers ---
  const handleSubmit = async () => {
    setError("");
    setResult(null);
    setLoading(true);

    try {
      let resp: Result;

      if (tab === "upload") {
        if (!file) { setError("Please select a file"); setLoading(false); return; }
        resp = await submitFile(file);
      } else if (tab === "paste") {
        if (!code.trim()) { setError("Please paste some code"); setLoading(false); return; }
        const blob = new File([code], "pasted-code.txt", { type: "text/plain" });
        resp = await submitFile(blob);
      } else {
        const repoErr = validateRepo(repoUrl);
        if (repoErr) { setError(repoErr); setLoading(false); return; }
        resp = await submitRepo(repoUrl);
      }

      setResult(resp);
      if (!resp.success && resp.error) {
        setError(resp.error.message);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Submission failed");
    } finally {
      setLoading(false);
    }
  };

  const tabClass = (t: Tab) =>
    `px-4 py-2 text-sm font-medium rounded-t-lg transition-colors ${
      tab === t
        ? "bg-white text-black border-b-2 border-black dark:bg-zinc-900 dark:text-white dark:border-white"
        : "text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-300"
    }`;

  return (
    <div className="w-full max-w-2xl mx-auto">
      {/* Tab bar */}
      <div className="flex gap-1 border-b border-zinc-200 dark:border-zinc-700">
        <button className={tabClass("upload")} onClick={() => setTab("upload")}>
          File Upload
        </button>
        <button className={tabClass("paste")} onClick={() => setTab("paste")}>
          Paste Code
        </button>
        <button className={tabClass("repo")} onClick={() => setTab("repo")}>
          GitHub Repo
        </button>
      </div>

      {/* Tab content */}
      <div className="bg-white dark:bg-zinc-900 border border-t-0 border-zinc-200 dark:border-zinc-700 rounded-b-lg p-6">
        {/* File upload with drag-and-drop */}
        {tab === "upload" && (
          <div
            onDragOver={onDragOver}
            onDragLeave={onDragLeave}
            onDrop={onDrop}
            onClick={() => fileRef.current?.click()}
            className={`border-2 border-dashed rounded-lg p-12 text-center cursor-pointer transition-colors ${
              dragging
                ? "border-blue-500 bg-blue-50 dark:bg-blue-950"
                : "border-zinc-300 hover:border-zinc-400 dark:border-zinc-600"
            }`}
          >
            <input
              ref={fileRef}
              type="file"
              className="hidden"
              onChange={(e) => setFile(e.target.files?.[0] || null)}
            />
            {file ? (
              <div>
                <p className="text-lg font-medium text-zinc-800 dark:text-zinc-200">{file.name}</p>
                <p className="text-sm text-zinc-500 mt-1">{(file.size / 1024).toFixed(1)} KB</p>
                <button
                  onClick={(e) => { e.stopPropagation(); setFile(null); }}
                  className="mt-3 text-sm text-red-500 hover:text-red-700"
                >
                  Remove
                </button>
              </div>
            ) : (
              <div>
                <p className="text-zinc-500 dark:text-zinc-400">
                  Drag and drop a file here, or click to browse
                </p>
                <p className="text-xs text-zinc-400 mt-2">Max 10MB</p>
              </div>
            )}
          </div>
        )}

        {/* Code paste textarea */}
        {tab === "paste" && (
          <textarea
            value={code}
            onChange={(e) => setCode(e.target.value)}
            placeholder="Paste your code here..."
            rows={12}
            className="w-full p-4 font-mono text-sm border border-zinc-300 dark:border-zinc-600 rounded-lg bg-zinc-50 dark:bg-zinc-800 text-zinc-800 dark:text-zinc-200 focus:outline-none focus:ring-2 focus:ring-blue-500 resize-y"
          />
        )}

        {/* GitHub repo URL input */}
        {tab === "repo" && (
          <div>
            <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-2">
              GitHub Repository URL
            </label>
            <input
              type="url"
              value={repoUrl}
              onChange={(e) => setRepoUrl(e.target.value)}
              placeholder="https://github.com/owner/repo"
              className="w-full p-3 border border-zinc-300 dark:border-zinc-600 rounded-lg bg-zinc-50 dark:bg-zinc-800 text-zinc-800 dark:text-zinc-200 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <p className="text-xs text-zinc-400 mt-2">Must be a public GitHub repository</p>
          </div>
        )}

        {/* Error display */}
        {error && (
          <div className="mt-4 p-3 bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg text-sm text-red-700 dark:text-red-300">
            {error}
          </div>
        )}

        {/* Success display */}
        {result?.success && (
          <div className="mt-4 p-3 bg-green-50 dark:bg-green-950 border border-green-200 dark:border-green-800 rounded-lg text-sm text-green-700 dark:text-green-300">
            <p className="font-medium">Submitted successfully</p>
            <pre className="mt-2 text-xs overflow-x-auto">{JSON.stringify(result.data, null, 2)}</pre>
          </div>
        )}

        {/* Submit button */}
        <button
          onClick={handleSubmit}
          disabled={loading}
          className="mt-6 w-full py-3 px-6 bg-black text-white rounded-lg font-medium hover:bg-zinc-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors dark:bg-white dark:text-black dark:hover:bg-zinc-200"
        >
          {loading ? "Analyzing..." : "Analyze"}
        </button>
      </div>
    </div>
  );
}
