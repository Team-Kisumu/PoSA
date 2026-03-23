"""
PoSA AI Evaluation Engine.

FastAPI server that receives code/text submissions from the Go backend
and returns evaluation results (score, issues, suggestions). This is the
scaffold — actual AI analysis will be integrated in subsequent issues.
"""

from fastapi import FastAPI, HTTPException, status
from pydantic import BaseModel, Field

app = FastAPI(
    title="PoSA AI Engine",
    description="AI evaluation engine for Proof-of-Skill submissions",
    version="0.1.0",
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


class EvaluationResponse(BaseModel):
    """Evaluation result returned to the Go backend."""

    score: int = Field(..., ge=0, le=100)
    issues: list[Issue] = []
    suggestions: list[str] = []


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
    Evaluate a code/text submission.

    Receives the submission payload from the Go backend, runs analysis,
    and returns a score (0-100), list of issues, and suggestions.

    Currently returns a placeholder response. Actual AI analysis
    (code quality, security, logic) will be integrated in Issue #8.
    """
    # Reject empty content for file submissions — the backend should
    # have caught this, but defense-in-depth.
    if req.submission_type == "file" and not req.content.strip():
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="file submission requires non-empty content",
        )

    # Placeholder evaluation — returns a fixed score with no issues.
    # This will be replaced with actual AI analysis in Phase 2.
    return EvaluationResponse(
        score=0,
        issues=[],
        suggestions=["AI evaluation not yet implemented"],
    )


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)
