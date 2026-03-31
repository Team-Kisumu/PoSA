// API client for the PoSA backend.

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface SubmitFileResponse {
  success: boolean;
  data?: {
    type: string;
    name: string;
    size: number;
    mime: string;
    message: string;
  };
  error?: { code: string; message: string };
}

export interface SubmitRepoResponse {
  success: boolean;
  data?: {
    type: string;
    name: string;
    size: number;
    message: string;
  };
  error?: { code: string; message: string };
}

export interface HealthResponse {
  success: boolean;
  data?: { status: string };
}

export async function checkHealth(): Promise<HealthResponse> {
  const resp = await fetch(`${API_URL}/health`);
  return resp.json();
}

export async function submitFile(file: File): Promise<SubmitFileResponse> {
  const form = new FormData();
  form.append("file", file);

  const resp = await fetch(`${API_URL}/api/submit`, {
    method: "POST",
    body: form,
  });
  return resp.json();
}

export async function submitRepo(repoUrl: string): Promise<SubmitRepoResponse> {
  const resp = await fetch(`${API_URL}/api/submit`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ repo: repoUrl }),
  });
  return resp.json();
}
