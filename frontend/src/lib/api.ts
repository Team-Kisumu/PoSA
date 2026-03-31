// API client for the PoSA backend.

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const AI_URL = process.env.NEXT_PUBLIC_AI_URL || "http://localhost:8000";

// --- Types ---

export interface APIResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: { code: string; message: string };
}

export interface SubmitData {
  type: string;
  name: string;
  size: number;
  mime: string;
  message: string;
}

export interface EvaluationIssue {
  issue_type: string;
  message: string;
  line: number | null;
}

export interface FileScore {
  path: string;
  score: number;
  issues: number;
  language: string | null;
}

export interface EvaluationResult {
  score: number;
  issues: EvaluationIssue[];
  suggestions: string[];
  files_analyzed?: number;
  file_scores?: FileScore[];
}

export interface MintResult {
  cid: string;
  tx_hash: string;
  score: number;
  submitter: string;
}

// --- API Calls ---

export async function checkHealth(): Promise<APIResponse> {
  const resp = await fetch(`${API_URL}/health`);
  return resp.json();
}

export async function submitFile(file: File): Promise<APIResponse<SubmitData>> {
  const form = new FormData();
  form.append("file", file);
  const resp = await fetch(`${API_URL}/api/submit`, { method: "POST", body: form });
  return resp.json();
}

export async function submitRepo(repoUrl: string): Promise<APIResponse<SubmitData>> {
  const resp = await fetch(`${API_URL}/api/submit`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ repo: repoUrl }),
  });
  return resp.json();
}

export async function evaluateFile(
  name: string,
  content: string,
  mime: string
): Promise<EvaluationResult> {
  const resp = await fetch(`${AI_URL}/evaluate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ submission_type: "file", name, content, mime }),
  });
  if (!resp.ok) throw new Error(`Evaluation failed: ${resp.status}`);
  return resp.json();
}

export async function evaluateRepo(repoUrl: string): Promise<EvaluationResult> {
  const resp = await fetch(`${AI_URL}/evaluate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ submission_type: "repo", name: repoUrl, content: "", mime: "" }),
  });
  if (!resp.ok) {
    const err = await resp.json().catch(() => ({ detail: `Evaluation failed: ${resp.status}` }));
    throw new Error(err.detail || `Evaluation failed: ${resp.status}`);
  }
  return resp.json();
}

export async function mintProof(
  cid: string,
  score: number,
  submitter: string
): Promise<MintResult> {
  // In the full pipeline this calls the backend which orchestrates
  // storage + blockchain. For now, simulate the response structure.
  const resp = await fetch(`${API_URL}/api/mint`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ cid, score, submitter }),
  });
  if (!resp.ok) throw new Error(`Minting failed: ${resp.status}`);
  return resp.json();
}

// --- Verification ---

export interface VerifyResult {
  valid: boolean;
  credential: {
    id: number;
    cid: string;
    score: number;
    submitter: string;
    timestamp: number;
  } | null;
  report: Record<string, unknown> | null;
}

export async function verifyCID(cid: string): Promise<VerifyResult> {
  // Fetch from backend verify endpoint.
  const resp = await fetch(`${API_URL}/api/verify/${encodeURIComponent(cid)}`);
  const data: APIResponse = await resp.json();

  if (!data.success) {
    return { valid: false, credential: null, report: null };
  }

  // Try to fetch the report from Lighthouse gateway.
  let report: Record<string, unknown> | null = null;
  try {
    const reportResp = await fetch(
      `https://gateway.lighthouse.storage/ipfs/${cid}`,
      { signal: AbortSignal.timeout(10000) }
    );
    if (reportResp.ok) {
      report = await reportResp.json();
    }
  } catch {
    // Report fetch failed — credential may still be valid on-chain.
  }

  return {
    valid: true,
    credential: data.data as VerifyResult["credential"],
    report,
  };
}
