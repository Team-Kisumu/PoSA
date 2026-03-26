"""
Tests for the PoSA code quality analyzer.

Covers language detection, rule matching for all supported languages,
scoring calculation, suggestion generation, and edge cases.
"""

from ai.analyzer import analyze, detect_language, is_known_code_file

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


def test_detect_c():
    assert detect_language("main.c") == "c"


def test_detect_java():
    assert detect_language("Main.java") == "java"


def test_detect_rust():
    assert detect_language("lib.rs") == "rust"


def test_detect_ruby():
    assert detect_language("app.rb") == "ruby"


def test_detect_php():
    assert detect_language("index.php") == "php"


def test_detect_shell():
    assert detect_language("deploy.sh") == "shell"


def test_detect_sql():
    assert detect_language("query.sql") == "sql"


def test_detect_css():
    assert detect_language("style.css") == "css"


def test_detect_config():
    assert detect_language("config.yaml") == "config"


def test_detect_terraform():
    assert detect_language("main.tf") == "terraform"


def test_detect_unsupported():
    assert detect_language("data.xyz") is None


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


# --- C/C++ Analysis Tests ---


def test_c_strcpy():
    """strcpy should be flagged as buffer overflow risk."""
    result = analyze("strcpy(dest, src);", "vuln.c")
    assert any("strcpy" in i["message"] for i in result["issues"])
    assert result["language"] == "c"


def test_c_system():
    """system() in C should be flagged."""
    result = analyze('system("ls -la");', "cmd.c")
    assert any("system()" in i["message"] for i in result["issues"])


def test_c_clean():
    """Clean C code should score 100."""
    result = analyze("int main() { return 0; }\n", "clean.c")
    assert result["score"] == 100


# --- Java Analysis Tests ---


def test_java_runtime_exec():
    """Runtime.exec should be flagged."""
    result = analyze('Runtime.getRuntime().exec("cmd");', "Cmd.java")
    assert any("Runtime.exec" in i["message"] for i in result["issues"])
    assert result["language"] == "java"


def test_java_sysout():
    """System.out.println should be flagged."""
    result = analyze('System.out.println("debug");', "Debug.java")
    assert any("System.out.print" in i["message"] for i in result["issues"])


# --- Rust Analysis Tests ---


def test_rust_unwrap():
    """unwrap() should be flagged."""
    result = analyze("let val = result.unwrap();", "main.rs")
    assert any("unwrap()" in i["message"] for i in result["issues"])
    assert result["language"] == "rust"


def test_rust_unsafe():
    """unsafe blocks should be flagged."""
    result = analyze("unsafe { ptr::read(p) }", "raw.rs")
    assert any("unsafe" in i["message"] for i in result["issues"])


# --- Ruby Analysis Tests ---


def test_ruby_eval():
    """eval in Ruby should be flagged."""
    result = analyze("eval(user_input)", "danger.rb")
    assert any("eval" in i["message"] for i in result["issues"])
    assert result["language"] == "ruby"


def test_ruby_system():
    """system() in Ruby should be flagged."""
    result = analyze('system("rm -rf /")', "bad.rb")
    assert any("system()" in i["message"] for i in result["issues"])


# --- PHP Analysis Tests ---


def test_php_eval():
    """eval in PHP should be flagged."""
    result = analyze("eval($code);", "bad.php")
    assert any("eval()" in i["message"] for i in result["issues"])
    assert result["language"] == "php"


def test_php_shell_exec():
    """shell_exec in PHP should be flagged."""
    result = analyze("shell_exec($cmd);", "cmd.php")
    assert any("shell_exec" in i["message"] for i in result["issues"])


# --- Shell Analysis Tests ---


def test_shell_eval():
    """eval in shell should be flagged."""
    result = analyze("eval $user_input", "bad.sh")
    assert any("eval" in i["message"] for i in result["issues"])
    assert result["language"] == "shell"


def test_shell_chmod_777():
    """chmod 777 should be flagged."""
    result = analyze("chmod 777 /var/www", "deploy.sh")
    assert any("chmod 777" in i["message"] for i in result["issues"])


# --- SQL Analysis Tests ---


def test_sql_select_star():
    """SELECT * should be flagged."""
    result = analyze("SELECT * FROM users;", "query.sql")
    assert any("SELECT *" in i["message"] for i in result["issues"])
    assert result["language"] == "sql"


def test_sql_grant_all():
    """GRANT ALL PRIVILEGES should be flagged."""
    result = analyze("GRANT ALL PRIVILEGES ON *.* TO 'user';", "perms.sql")
    assert any("GRANT ALL" in i["message"] for i in result["issues"])


# --- CSS Analysis Tests ---


def test_css_important():
    """!important should be flagged."""
    result = analyze("color: red !important;", "style.css")
    assert any("!important" in i["message"] for i in result["issues"])
    assert result["language"] == "css"


# --- Config Analysis Tests ---


def test_config_hardcoded_secret():
    """Hardcoded secrets in config should be flagged."""
    result = analyze('api_key: "sk-1234567890abcdef"', "config.yaml")
    assert any("secret" in i["message"].lower() for i in result["issues"])
    assert result["language"] == "config"


def test_config_clean():
    """Config without secrets should score 100."""
    result = analyze("port: 8080\nhost: api.example.com\n", "config.yaml")
    assert result["score"] == 100


# --- Terraform Analysis Tests ---


def test_terraform_open_cidr():
    """0.0.0.0/0 in security groups should be flagged."""
    result = analyze('cidr_blocks = ["0.0.0.0/0"]', "main.tf")
    assert any("0.0.0.0/0" in i["message"] for i in result["issues"])
    assert result["language"] == "terraform"


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
    """Truly unknown file types return score 0 with 'not supported'."""
    result = analyze("some random content", "data.xyz")
    assert result["score"] == 0
    assert result["language"] is None
    assert any("not supported" in s.lower() for s in result["suggestions"])


def test_is_known_code_file():
    """is_known_code_file should recognize common code extensions."""
    assert is_known_code_file("main.c") is True
    assert is_known_code_file("App.java") is True
    assert is_known_code_file("lib.rs") is True
    assert is_known_code_file("script.rb") is True
    assert is_known_code_file("style.css") is True
    assert is_known_code_file("query.sql") is True
    assert is_known_code_file("main.go") is True
    assert is_known_code_file("data.xyz") is False
    assert is_known_code_file("README.md") is False


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
