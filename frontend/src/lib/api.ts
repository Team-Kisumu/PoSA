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

export interface EvaluationResult {
  score: number;
  issues: EvaluationIssue[];
  suggestions: string[];
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
