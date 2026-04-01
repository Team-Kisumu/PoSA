"""
PoSA AI Evaluation Engine.

FastAPI server that receives code/text submissions from the Go backend
and returns evaluation results (score, issues, suggestions). Supports:
  - Code analysis: Go, Python, JavaScript (pattern-based static analysis)
  - Writing evaluation: Markdown, text, RST, HTML (quality and readability)
"""

from fastapi import FastAPI, HTTPException, status
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field

from ai.analyzer import analyze
from ai.repo_analyzer import analyze_repo

app = FastAPI(
    title="PoSA AI Engine",
    description="AI evaluation engine for Proof-of-Skill submissions",
    version="0.2.0",
)

# Allow cross-origin requests from the frontend.
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)


# --- Request/Response Models ---


class EvaluationRequest(BaseModel):
    """Payload sent by the Go backend for evaluation."""

    # Submission type: "file" for uploaded files, "repo" for GitHub repo links.
    submission_type: str = Field(..., pattern="^(file|repo)$")
    # Original filename or repo URL.
    name: str = Field(..., min_length=1, max_length=500)
    # File content (for file submissions) or empty string (for repo submissions).
    content: str = Field(default="")
    # Detected MIME type from the backend validation layer.
    mime: str = Field(default="")


class Issue(BaseModel):
    """A single issue found during evaluation."""

    issue_type: str  # "security", "quality", "logic", "style"
    message: str
    line: int | None = None  # Line number, if applicable.


class FileScore(BaseModel):
    """Per-file score in a repo evaluation."""

    path: str
    score: int
    issues: int
    language: str | None = None


class EvaluationResponse(BaseModel):
    """Evaluation result returned to the Go backend."""

    score: int = Field(..., ge=0, le=100)
    issues: list[Issue] = []
    suggestions: list[str] = []
    files_analyzed: int | None = None
    file_scores: list[FileScore] | None = None


# --- Endpoints ---


@app.get("/health")
def health():
    """Health check endpoint for the AI engine."""
    return {"status": "ok"}


@app.post(
    "/evaluate",
    response_model=EvaluationResponse,
    status_code=status.HTTP_200_OK,
)
def evaluate(req: EvaluationRequest):
    """
    Evaluate a code or writing submission.

    Routes to the appropriate evaluator based on file type:
    - Code files (.go, .py, .js/.ts): static analysis for security,
      quality, and style issues.
    - Writing files (.md, .txt, .rst, .html): quality, structure,
      and readability evaluation.
    - Unsupported types: score 0 with suggestion.
    """
    # Reject empty content for file submissions — the backend should
    # have caught this, but defense-in-depth.
    if req.submission_type == "file" and not req.content.strip():
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="file submission requires non-empty content",
        )

    # Repo submissions: fetch files from GitHub and analyze.
    if req.submission_type == "repo":
        try:
            result = analyze_repo(req.name)
        except ValueError as e:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=str(e),
            )
        except Exception as e:
            raise HTTPException(
                status_code=status.HTTP_502_BAD_GATEWAY,
                detail=f"failed to fetch repository: {e}",
            )

        issues = [
            Issue(
                issue_type=i["issue_type"],
                message=i["message"],
                line=i.get("line"),
            )
            for i in result["issues"]
        ]
        file_scores = [FileScore(**fs) for fs in result.get("file_scores", [])]
        return EvaluationResponse(
            score=result["score"],
            issues=issues,
            suggestions=result["suggestions"],
            files_analyzed=result.get("files_analyzed", 0),
            file_scores=file_scores if file_scores else None,
        )

    # Run static analysis on the submitted code.
    result = analyze(req.content, req.name)

    # Convert analyzer output to response model format.
    # The analyzer returns severity in issues (used for scoring) but
    # the API response model doesn't include it — strip before returning.
    issues = [
        Issue(
            issue_type=i["issue_type"],
            message=i["message"],
            line=i.get("line"),
        )
        for i in result["issues"]
    ]

    return EvaluationResponse(
        score=result["score"],
        issues=issues,
        suggestions=result["suggestions"],
    )


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)
