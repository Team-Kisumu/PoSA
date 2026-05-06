// API client for the PoSA backend.
// In development: NEXT_PUBLIC_API_URL=http://localhost:8080, NEXT_PUBLIC_AI_URL=http://localhost:8000
// In production: both empty — nginx proxies /api/* to backend, /ai/* to AI engine on same origin.

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
const AI_URL = process.env.NEXT_PUBLIC_AI_URL ?? "http://localhost:8000";

// In production (empty URL), rewrite paths for the nginx proxy:
//   backend: /health, /api/* stay as-is (nginx routes them)
//   AI engine: /evaluate -> /ai/evaluate (nginx strips /ai prefix)
const aiPrefix = AI_URL === "" ? "/ai" : AI_URL;

// --- Types ---

// Read the CSRF token from the cookie for state-changing requests.
function getCsrfToken(): string {
  if (typeof document === "undefined") return "";
  const match = document.cookie.match(/csrf_token=([^;]+)/);
  return match ? match[1] : "";
}

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
  const resp = await fetch(`${API_URL}/api/submit`, {
    method: "POST",
    body: form,
    headers: { "X-CSRF-Token": getCsrfToken() },
    credentials: "include",
  });
  return resp.json();
}

export async function submitRepo(repoUrl: string): Promise<APIResponse<SubmitData>> {
  const resp = await fetch(`${API_URL}/api/submit`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "X-CSRF-Token": getCsrfToken() },
    body: JSON.stringify({ repo: repoUrl }),
    credentials: "include",
  });
  return resp.json();
}

export async function evaluateFile(
  name: string,
  content: string,
  mime: string
): Promise<EvaluationResult> {
  const resp = await fetch(`${aiPrefix}/evaluate`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "X-CSRF-Token": getCsrfToken() },
    body: JSON.stringify({ submission_type: "file", name, content, mime }),
    credentials: "include",
  });
  if (!resp.ok) throw new Error(`Evaluation failed: ${resp.status}`);
  return resp.json();
}

export async function evaluateRepo(repoUrl: string): Promise<EvaluationResult> {
  const resp = await fetch(`${aiPrefix}/evaluate`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "X-CSRF-Token": getCsrfToken() },
    body: JSON.stringify({ submission_type: "repo", name: repoUrl, content: "", mime: "" }),
    credentials: "include",
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
    headers: { "Content-Type": "application/json", "X-CSRF-Token": getCsrfToken() },
    body: JSON.stringify({ cid, score, submitter }),
    credentials: "include",
  });
  if (!resp.ok) throw new Error(`Minting failed: ${resp.status}`);
  return resp.json();
}

// --- Verification ---

export interface UserProfile {
  id: number;
  github_id: number;
  username: string;
  avatar_url: string;
  email: string;
  role: string;
}

export async function getMe(): Promise<UserProfile | null> {
  try {
    const resp = await fetch(`${API_URL}/auth/me`, { credentials: "include" });
    if (!resp.ok) return null;
    const data: APIResponse<UserProfile> = await resp.json();
    return data.success ? (data.data as UserProfile) : null;
  } catch {
    return null;
  }
}

export function getLoginUrl(): string {
  return `${API_URL}/auth/github`;
}

export async function logout(): Promise<void> {
  await fetch(`${API_URL}/auth/logout`, {
    method: "POST",
    headers: { "X-CSRF-Token": getCsrfToken() },
    credentials: "include",
  });
}


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
