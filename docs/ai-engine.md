# AI Engine

> **Status:** Full analysis implemented (Phase 2)

The AI engine evaluates user-submitted work using pattern-based analysis:

- **Code analysis:** 25 languages with language-specific + cross-language rules
- **Writing evaluation:** Quality, structure, and readability scoring for text documents

## Technology

- **Language:** Python 3.10+
- **Framework:** FastAPI 0.115+
- **Server:** Uvicorn
- **Testing:** pytest + httpx (via FastAPI TestClient)
- **Dependencies:** Defined in `ai/requirements.txt`

## Directory Structure

```md
ai/
├── __init__.py              # Package marker
├── evaluator.py             # FastAPI app with health and evaluation endpoints
├── analyzer.py              # Router: detects file type, delegates to evaluator
├── writing.py               # Writing quality evaluator (structure, readability, style)
├── requirements.txt         # Direct Python dependencies
├── rules/
│   ├── __init__.py          # Package marker
│   ├── patterns.py          # Extension map, language rules registry, Go/Python/JS rules
│   ├── compiled_langs.py    # C/C++, Java, C#, Ruby, Rust, PHP, Swift, Shell rules
│   ├── other_langs.py       # Perl, R, Lua, Dart, Elixir, Erlang, Haskell, Clojure,
│   │                        # Scala, Kotlin, SQL, CSS, Config, Terraform rules
│   └── quality_rules.py     # Cross-language: formatting, incompleteness, logic,
│                            # deprecated, obsolete rules
└── tests/
    ├── __init__.py          # Package marker
    ├── test_analyzer.py     # Code analyzer tests (76 cases)
    ├── test_writing.py      # Writing evaluator tests (29 cases)
    ├── test_quality_rules.py # Cross-language rule tests (26 cases)
    └── test_evaluator.py    # Endpoint integration tests (19 cases)
```

## Analysis Architecture

### Pipeline

```mmd
File Upload → File Type Detection → Code Analyzer / Writing Evaluator → Scoring → Response
```

The analyzer routes submissions based on file extension:

- 67 code extensions → language-specific + cross-language rules
- Writing extensions (.md, .txt, .rst, .html, .tex, .adoc) → writing quality checks
- Unknown extensions → score 0 with suggestion

### Rule Categories

Every code file is checked against **8 categories** of rules:

| Category | Issue Type | What It Detects |
|---|---|---|
| Security | `security` | Command injection, SQL injection, XSS, deserialization, buffer overflows |
| Quality | `quality` | Error handling, TODO comments, unsafe patterns, empty catch blocks |
| Style | `style` | Debug output, logging recommendations, naming conventions |
| Formatting | `formatting` | Mixed indentation, trailing whitespace, long lines, blank line runs |
| Incompleteness | `incompleteness` | Stubs, placeholders, NotImplementedError, HACK comments, debug output |
| Logic | `logic` | Assignment in conditionals, unreachable code, self-comparison, division by zero |
| Deprecated | `deprecated` | Removed APIs (gets, distutils, mysql_*, auto_ptr, ioutil, finalize) |
| Obsolete | `obsolete` | Old copyright years, EOL versions, legacy tools (Grunt, Bower), deprecated React lifecycle |

### Supported Languages (25)

| Language | Extensions | Language-Specific Rules |
|---|---|---|
| Go | `.go` | 10 (exec.Command, SQL injection, nil handler, panic, fmt.Println, etc.) |
| Python | `.py` | 12 (eval, exec, os.system, subprocess shell, pickle, yaml.load, etc.) |
| JavaScript/TS | `.js`, `.jsx`, `.ts`, `.tsx`, `.mjs`, `.cjs` | 11 (eval, innerHTML, document.write, child_process, var, ==, etc.) |
| C/C++ | `.c`, `.h`, `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.hh` | 8 (gets, strcpy, strcat, sprintf, system, malloc, printf format) |
| Java | `.java` | 8 (Runtime.exec, SQL concat, ObjectInputStream, printStackTrace, Random) |
| C# | `.cs`, `.fs`, `.fsx`, `.vb` | 6 (Process.Start, SQL concat, BinaryFormatter, empty catch) |
| Ruby | `.rb`, `.erb`, `.rake`, `.gemspec` | 7 (eval, system, backticks, Marshal.load, bare rescue) |
| Rust | `.rs` | 6 (unsafe, unwrap, expect, Command::new, println!) |
| PHP | `.php` | 8 (eval, exec, shell_exec, system, SQL interpolation, unserialize, XSS) |
| Swift | `.swift`, `.m` | 5 (try!, force unwrap, Process, print) |
| Shell | `.sh`, `.bash`, `.zsh`, `.fish` | 6 (eval, chmod 777, curl\|bash, unquoted vars) |
| Perl | `.pl`, `.pm` | 4 (eval, system, backticks) |
| R | `.r`, `.R` | 4 (eval, system, system2) |
| Lua | `.lua` | 4 (loadstring, os.execute, io.popen) |
| Dart | `.dart` | 4 (Process.run, force unwrap, print) |
| Elixir | `.ex`, `.exs` | 5 (Code.eval_string, System.cmd, binary_to_term, IO.puts) |
| Erlang | `.erl`, `.hrl` | 3 (os:cmd, binary_to_term) |
| Haskell | `.hs`, `.lhs` | 4 (unsafePerformIO, head, system) |
| Clojure | `.clj`, `.cljs`, `.cljc`, `.edn` | 3 (eval, sh) |
| Scala | `.scala`, `.sc` | 4 (Runtime.exec, sys.process, println) |
| Kotlin | `.kt`, `.kts`, `.groovy`, `.gradle` | 4 (Runtime.exec, !!, println) |
| SQL | `.sql` | 5 (GRANT ALL, SELECT *, DROP without IF EXISTS, hardcoded passwords) |
| CSS | `.css`, `.scss`, `.sass`, `.less` | 4 (expression(), !important, universal selector) |
| Config | `.json`, `.yaml`, `.yml`, `.toml`, `.xml`, `.ini`, `.cfg`, `.conf`, `.env`, `.properties`, `.csv`, `.ipynb` | 3 (hardcoded secrets, localhost addresses) |
| Terraform | `.tf`, `.hcl`, `.dockerfile` | 4 (0.0.0.0/0 CIDR, hardcoded secrets, encryption disabled) |

**Plus 34 cross-language rules** applied to every code file (formatting, incompleteness, logic, deprecated, obsolete).

### Writing Evaluation

| Check | What It Detects |
|---|---|
| Content length | Very short (<20 words) or short (<50 words) submissions |
| Structure | Wall-of-text (no paragraph breaks), missing headings |
| Sentence quality | Average length >30 words, individual sentences >40 words |
| Passive voice | "to be + past participle" patterns |
| Weasel words | Vague language (very, really, basically, etc.) |
| Repeated phrases | Same word 3+ times in a line |

### Scoring System

| Severity | Point Deduction | Examples |
|---|---|---|
| High (-15) | Security vulnerabilities, critical bugs | eval(), SQL injection, division by zero, buffer overflow |
| Medium (-8) | Quality concerns, incomplete code | panic(), bare except, stubs, deprecated APIs, unsafe blocks |
| Low (-3) | Style and formatting issues | TODO comments, trailing whitespace, debug output, !important |

Score = max(0, 100 - sum of deductions). Clean code scores 100.

## API Reference

### `GET /health`

**Response:** `200 OK`

```json
{"status": "ok"}
```

### `POST /evaluate`

**Request:**

```json
{
  "submission_type": "file",
  "name": "main.go",
  "content": "package main\n\nimport \"os/exec\"\n\nfunc run(cmd string) {\n\texec.Command(cmd)\n}\n",
  "mime": "text/plain"
}
```

**Response:** `200 OK`

```json
{
  "score": 85,
  "issues": [
    {
      "issue_type": "security",
      "message": "Use of exec.Command — verify input is sanitized to prevent command injection",
      "line": 6
    }
  ],
  "suggestions": [
    "Security issues detected — review and fix before deployment"
  ]
}
```

**Error responses:**

| Status | Condition |
|---|---|
| `400` | File submission with empty content |
| `422` | Invalid submission_type, missing required fields, name too long, malformed JSON |

## Running

```bash
cd ai
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt

# Start the server
python evaluator.py
# -> Uvicorn running on http://0.0.0.0:8000

# Interactive docs
# Swagger UI: http://localhost:8000/docs
# ReDoc: http://localhost:8000/redoc
```

## Testing

```bash
# From project root (with venv activated)
PYTHONPATH=. pytest ai/tests/ -v

# 150 tests, ~1s
```

### Test Summary

| File | Tests | Covers |
|---|---|---|
| `test_analyzer.py` | 76 | Language detection (18), Go/Python/JS rules (27), C/Java/Rust/Ruby/PHP/Shell/SQL/CSS/Config/Terraform rules (20), scoring (4), suggestions (3), edge cases (4) |
| `test_writing.py` | 29 | File type detection (8), quality scoring (4), structure (4), passive voice (2), weasel words (2), sentence length (1), suggestions (3), analyzer routing (3), edge cases (2) |
| `test_quality_rules.py` | 26 | Formatting (3), incompleteness (4), logic errors (4), deprecated APIs (6), obsolete patterns (4), suggestion generation (5) |
| `test_evaluator.py` | 19 | Health (2), endpoint integration (7), validation (8), response structure (2) |
| **Total** | **150** | |

## Next Steps

- Future: Plagiarism detection for writing submissions
