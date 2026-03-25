"""
Tests for the PoSA code quality analyzer.

Covers language detection, rule matching for Go/Python/JavaScript,
scoring calculation, suggestion generation, and edge cases.
"""

from ai.analyzer import analyze, detect_language

# --- Language Detection Tests ---


def test_detect_go():
    assert detect_language("main.go") == "go"


def test_detect_python():
    assert detect_language("script.py") == "python"


def test_detect_javascript():
    assert detect_language("app.js") == "javascript"


def test_detect_typescript():
    """TypeScript files are analyzed with JavaScript rules."""
    assert detect_language("app.ts") == "javascript"


def test_detect_jsx():
    assert detect_language("component.jsx") == "javascript"


def test_detect_tsx():
    assert detect_language("component.tsx") == "javascript"


def test_detect_unsupported():
    assert detect_language("style.css") is None


def test_detect_no_extension():
    assert detect_language("Makefile") is None


# --- Go Analysis Tests ---


def test_go_clean_code():
    """Clean Go code should score 100 with no issues."""
    code = """package main

import "net/http"

func handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
}
"""
    result = analyze(code, "handler.go")
    assert result["score"] == 100
    assert result["issues"] == []
    assert result["language"] == "go"


def test_go_exec_command():
    """exec.Command should be flagged as a security issue."""
    code = """package main

import "os/exec"

func run(cmd string) {
    exec.Command(cmd)
}
"""
    result = analyze(code, "runner.go")
    assert result["score"] < 100
    assert any(i["issue_type"] == "security" for i in result["issues"])
    assert any("exec.Command" in i["message"] for i in result["issues"])


def test_go_sql_injection():
    """String concatenation in SQL queries should be flagged."""
    code = """package main

func query(db *sql.DB, name string) {
    db.Query("SELECT * FROM users WHERE name=" + name)
}
"""
    # The pattern matches sql.Query with + concatenation.
    result = analyze(code, "db.go")
    security_issues = [i for i in result["issues"] if i["issue_type"] == "security"]
    assert len(security_issues) > 0


def test_go_panic():
    """Use of panic should be flagged as a quality issue."""
    code = """package main

func init() {
    panic("something went wrong")
}
"""
    result = analyze(code, "main.go")
    assert any(i["issue_type"] == "quality" for i in result["issues"])
    assert any("panic" in i["message"] for i in result["issues"])


def test_go_todo_comment():
    """TODO comments should be flagged."""
    code = """package main

// TODO: fix this later
func placeholder() {}
"""
    result = analyze(code, "main.go")
    assert any("TODO" in i["message"] for i in result["issues"])


def test_go_fmt_println():
    """fmt.Println should suggest structured logging."""
    code = """package main

import "fmt"

func debug() {
    fmt.Println("debug output")
}
"""
    result = analyze(code, "debug.go")
    assert any("fmt.Println" in i["message"] for i in result["issues"])


def test_go_nil_handler():
    """http.ListenAndServe with nil handler should be flagged."""
    code = """package main

import "net/http"

func main() {
    http.ListenAndServe(":8080", nil)
}
"""
    result = analyze(code, "server.go")
    assert any("default ServeMux" in i["message"] for i in result["issues"])


# --- Python Analysis Tests ---


def test_python_clean_code():
    """Clean Python code should score 100."""
    code = """import logging

logger = logging.getLogger(__name__)

def greet(name: str) -> str:
    logger.info("Greeting %s", name)
    return f"Hello, {name}"
"""
    result = analyze(code, "greet.py")
    assert result["score"] == 100
    assert result["issues"] == []
    assert result["language"] == "python"


def test_python_eval():
    """eval() should be flagged as a security issue."""
    code = """def compute(expr):
    return eval(expr)
"""
    result = analyze(code, "compute.py")
    assert any("eval()" in i["message"] for i in result["issues"])
    assert any(i["issue_type"] == "security" for i in result["issues"])


def test_python_exec():
    """exec() should be flagged as a security issue."""
    code = """def run_code(code):
    exec(code)
"""
    result = analyze(code, "runner.py")
    assert any("exec()" in i["message"] for i in result["issues"])


def test_python_os_system():
    """os.system() should be flagged."""
    code = """import os

def run(cmd):
    os.system(cmd)
"""
    result = analyze(code, "shell.py")
    assert any("os.system" in i["message"] for i in result["issues"])


def test_python_subprocess_shell():
    """subprocess with shell=True should be flagged."""
    code = """import subprocess

def run(cmd):
    subprocess.run(cmd, shell=True)
"""
    result = analyze(code, "runner.py")
    assert any("shell=True" in i["message"] for i in result["issues"])


def test_python_pickle():
    """pickle.load/loads should be flagged."""
    code = """import pickle

def load_data(data):
    return pickle.loads(data)
"""
    result = analyze(code, "loader.py")
    assert any("pickle" in i["message"] for i in result["issues"])


def test_python_bare_except():
    """Bare except clause should be flagged."""
    code = """def risky():
    try:
        do_something()
    except:
        pass
"""
    result = analyze(code, "handler.py")
    assert any("Bare except" in i["message"] for i in result["issues"])


def test_python_yaml_load():
    """yaml.load without safe Loader should be flagged."""
    code = """import yaml

def parse(data):
    return yaml.load(data)
"""
    result = analyze(code, "config.py")
    assert any("yaml.load" in i["message"] for i in result["issues"])


def test_python_wildcard_import():
    """Wildcard imports should be flagged."""
    code = """from os import *

def list_files():
    return listdir(".")
"""
    result = analyze(code, "files.py")
    assert any("Wildcard import" in i["message"] for i in result["issues"])


def test_python_print():
    """print() should suggest using logging."""
    code = """def debug():
    print("debug info")
"""
    result = analyze(code, "debug.py")
    assert any("print()" in i["message"] for i in result["issues"])


# --- JavaScript Analysis Tests ---


def test_js_clean_code():
    """Clean JavaScript code should score 100."""
    code = """const greet = (name) => {
    return `Hello, ${name}`;
};

export default greet;
"""
    result = analyze(code, "greet.js")
    assert result["score"] == 100
    assert result["issues"] == []
    assert result["language"] == "javascript"


def test_js_eval():
    """eval() should be flagged."""
    code = """function compute(expr) {
    return eval(expr);
}
"""
    result = analyze(code, "compute.js")
    assert any("eval()" in i["message"] for i in result["issues"])


def test_js_innerhtml():
    """innerHTML assignment should be flagged as XSS risk."""
    code = """function render(data) {
    document.getElementById("output").innerHTML = data;
}
"""
    result = analyze(code, "render.js")
    assert any("innerHTML" in i["message"] for i in result["issues"])


def test_js_document_write():
    """document.write should be flagged."""
    code = """function output(text) {
    document.write(text);
}
"""
    result = analyze(code, "output.js")
    assert any("document.write" in i["message"] for i in result["issues"])


def test_js_var():
    """var should suggest let/const."""
    code = """function count() {
    var x = 0;
    return x;
}
"""
    result = analyze(code, "counter.js")
    assert any("var" in i["message"] for i in result["issues"])


def test_js_loose_equality():
    """Loose equality (==) should suggest strict equality (===)."""
    code = """function check(a, b) {
    return a == b;
}
"""
    result = analyze(code, "check.js")
    assert any("==" in i["message"] for i in result["issues"])


def test_js_console_log():
    """console.log should be flagged."""
    code = """function debug() {
    console.log("test");
}
"""
    result = analyze(code, "debug.js")
    assert any("console.log" in i["message"] for i in result["issues"])


def test_js_new_function():
    """new Function() should be flagged as eval equivalent."""
    code = """const fn = new Function("return 42");
"""
    result = analyze(code, "dynamic.js")
    assert any("new Function" in i["message"] for i in result["issues"])


def test_js_child_process():
    """child_process usage should be flagged."""
    code = """const { exec } = require("child_process");
child_process.exec("ls -la");
"""
    result = analyze(code, "shell.js")
    assert any("child_process" in i["message"] for i in result["issues"])


def test_js_settimeout_string():
    """setTimeout with string argument should be flagged."""
    code = """setTimeout("alert('xss')", 1000);
"""
    result = analyze(code, "timer.js")
    assert any("setTimeout" in i["message"] for i in result["issues"])


# --- Scoring Tests ---


def test_score_no_issues():
    """Clean code should score 100."""
    result = analyze("const x = 42;\n", "clean.js")
    assert result["score"] == 100


def test_score_high_severity():
    """A single high-severity issue should deduct 15 points."""
    code = "eval('dangerous');\n"
    result = analyze(code, "bad.js")
    assert result["score"] == 85  # 100 - 15


def test_score_multiple_issues():
    """Multiple issues should accumulate deductions."""
    code = """var x = eval("code");
console.log(x);
"""
    result = analyze(code, "multi.js")
    # eval (high=15) + var (low=3) + console.log (low=3) = 21 deducted
    assert result["score"] == 79


def test_score_floor_at_zero():
    """Score should never go below 0."""
    # Stack many high-severity issues to exceed 100 points of deductions.
    code = "\n".join([f'eval("line{i}");' for i in range(10)])
    result = analyze(code, "terrible.js")
    assert result["score"] == 0


# --- Suggestion Tests ---


def test_suggestions_security():
    """Security issues should produce a security suggestion."""
    code = "eval('x');\n"
    result = analyze(code, "bad.js")
    assert any("Security" in s for s in result["suggestions"])


def test_suggestions_quality():
    """Quality issues should produce a quality suggestion."""
    code = "var x = 1;\n"
    result = analyze(code, "old.js")
    assert any("quality" in s.lower() for s in result["suggestions"])


def test_suggestions_clean():
    """Clean code should get a positive suggestion."""
    result = analyze("const x = 42;\n", "clean.js")
    assert any("clean" in s.lower() for s in result["suggestions"])


# --- Edge Cases ---


def test_unsupported_language():
    """Unsupported file types return score 0 with a suggestion."""
    result = analyze("body { color: red; }", "style.css")
    assert result["score"] == 0
    assert result["language"] is None
    assert any("not supported" in s.lower() for s in result["suggestions"])


def test_empty_content():
    """Empty content returns clean result for supported language."""
    result = analyze("", "empty.go")
    assert result["score"] == 100
    assert result["issues"] == []


def test_line_numbers():
    """Issues should report correct line numbers."""
    code = """package main

import "os/exec"

func run() {
    exec.Command("ls")
}
"""
    result = analyze(code, "runner.go")
    exec_issues = [i for i in result["issues"] if "exec.Command" in i["message"]]
    assert len(exec_issues) == 1
    assert exec_issues[0]["line"] == 6


def test_deduplication():
    """Same issue on the same line should not be reported twice."""
    code = "eval('x');\n"
    result = analyze(code, "dup.js")
    eval_issues = [i for i in result["issues"] if "eval()" in i["message"]]
    assert len(eval_issues) == 1
