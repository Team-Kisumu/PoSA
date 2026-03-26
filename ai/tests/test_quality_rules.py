"""
Tests for cross-language quality rules.

Covers formatting, incompleteness, wrong code/logic errors,
deprecated APIs, and obsolete patterns.
"""

from ai.analyzer import analyze

# --- Formatting Rules ---


def test_trailing_whitespace():
    """Trailing whitespace should be flagged."""
    result = analyze("int x = 1;   \n", "main.c")
    assert any("Trailing whitespace" in i["message"] for i in result["issues"])


def test_long_line():
    """Lines over 200 chars should be flagged."""
    code = "x = " + "a" * 200 + ";\n"
    result = analyze(code, "long.py")
    assert any("200 characters" in i["message"] for i in result["issues"])


def test_mixed_indentation():
    """Mixed tabs and spaces should be flagged."""
    code = "\t code_here();\n"
    result = analyze(code, "mixed.js")
    assert any("Mixed tabs" in i["message"] for i in result["issues"])


# --- Incompleteness Rules ---


def test_not_implemented_error():
    """NotImplementedError should be flagged as incomplete."""
    code = "def handler():\n    raise NotImplementedError\n"
    result = analyze(code, "stub.py")
    assert any("incompleteness" == i["issue_type"] for i in result["issues"])


def test_placeholder_variable():
    """Placeholder variable names should be flagged."""
    code = "placeholder = get_data()\n"
    result = analyze(code, "temp.py")
    assert any("Placeholder" in i["message"] for i in result["issues"])


def test_hack_comment():
    """HACK/WORKAROUND comments should be flagged."""
    code = "// WORKAROUND: this fixes the race condition\nfunc fix() {}\n"
    result = analyze(code, "fix.go")
    assert any("Workaround" in i["message"] for i in result["issues"])


def test_debug_output():
    """Debug/test output left in code should be flagged."""
    code = 'console.log("debug test output");\n'
    result = analyze(code, "app.js")
    assert any("incompleteness" == i["issue_type"] or "style" == i["issue_type"] for i in result["issues"])


# --- Wrong Code / Logic Error Rules ---


def test_assignment_in_conditional():
    """Assignment in if condition (= instead of ==) should be flagged."""
    code = "if (x = y) { doSomething(); }\n"
    result = analyze(code, "bug.js")
    assert any("logic" == i["issue_type"] for i in result["issues"])


def test_constant_condition():
    """Constant condition in control flow should be flagged."""
    code = "while (true) { process(); }\n"
    result = analyze(code, "loop.js")
    assert any("Constant condition" in i["message"] for i in result["issues"])


def test_self_comparison():
    """Comparing a variable to itself should be flagged."""
    code = "if (count == count) { return; }\n"
    result = analyze(code, "check.js")
    assert any("Comparing variable to itself" in i["message"] for i in result["issues"])


def test_division_by_zero():
    """Division by zero should be flagged."""
    code = "result = total / 0;\n"
    result = analyze(code, "math.js")
    assert any("Division by zero" in i["message"] for i in result["issues"])


# --- Deprecated API Rules ---


def test_deprecated_ioutil_go():
    """Go ioutil package should be flagged as deprecated."""
    code = "data, err := ioutil.ReadAll(r)\n"
    result = analyze(code, "reader.go")
    assert any("ioutil" in i["message"] for i in result["issues"])
    assert any("deprecated" == i["issue_type"] for i in result["issues"])


def test_deprecated_distutils_python():
    """Python distutils should be flagged as deprecated."""
    code = "from distutils.core import setup\n"
    result = analyze(code, "setup.py")
    assert any("distutils" in i["message"] for i in result["issues"])


def test_deprecated_proto_js():
    """__proto__ assignment should be flagged as deprecated."""
    code = "obj.__proto__ = newProto;\n"
    result = analyze(code, "legacy.js")
    assert any("__proto__" in i["message"] for i in result["issues"])


def test_deprecated_auto_ptr_cpp():
    """auto_ptr should be flagged as deprecated."""
    code = "std::auto_ptr<int> p(new int(42));\n"
    result = analyze(code, "old.cpp")
    assert any("auto_ptr" in i["message"] for i in result["issues"])


def test_deprecated_mysql_php():
    """mysql_* functions should be flagged as deprecated."""
    code = "mysql_connect('localhost', 'user', 'pass');\n"
    result = analyze(code, "db.php")
    assert any("mysql_" in i["message"] for i in result["issues"])


def test_deprecated_finalize_java():
    """finalize() should be flagged as deprecated."""
    code = "@Override protected void finalize() { super.finalize(); }\n"
    result = analyze(code, "Old.java")
    assert any("finalize" in i["message"] for i in result["issues"])


# --- Obsolete Codebase Rules ---


def test_obsolete_copyright():
    """Old copyright years should be flagged."""
    code = "// Copyright 2015 Company Inc.\npackage main\n"
    result = analyze(code, "main.go")
    assert any("copyright" in i["message"].lower() for i in result["issues"])


def test_obsolete_react_lifecycle():
    """Deprecated React lifecycle methods should be flagged."""
    code = "componentWillMount() { this.init(); }\n"
    result = analyze(code, "App.jsx")
    assert any("React lifecycle" in i["message"] for i in result["issues"])


def test_obsolete_jquery():
    """jQuery AJAX should be flagged as obsolete."""
    code = '$.ajax({ url: "/api/data" });\n'
    result = analyze(code, "app.js")
    assert any("jQuery" in i["message"] for i in result["issues"])


def test_obsolete_grunt():
    """Legacy build tool references should be flagged."""
    code = '"devDependencies": { "gruntfile": "^1.0" }\n'
    result = analyze(code, "package.json")
    assert any("Legacy build tool" in i["message"] for i in result["issues"])


# --- Suggestion Generation ---


def test_suggestion_formatting():
    """Formatting issues should produce a formatting suggestion."""
    result = analyze("int x = 1;   \n", "main.c")
    assert any("formatter" in s.lower() or "formatting" in s.lower() for s in result["suggestions"])


def test_suggestion_incompleteness():
    """Incomplete code should produce an incompleteness suggestion."""
    code = "def handler():\n    raise NotImplementedError\n"
    result = analyze(code, "stub.py")
    assert any("incomplete" in s.lower() or "finish" in s.lower() for s in result["suggestions"])


def test_suggestion_logic():
    """Logic errors should produce a logic suggestion."""
    code = "if (x = y) { doSomething(); }\n"
    result = analyze(code, "bug.js")
    assert any("logic" in s.lower() for s in result["suggestions"])


def test_suggestion_deprecated():
    """Deprecated APIs should produce a deprecated suggestion."""
    code = "data, err := ioutil.ReadAll(r)\n"
    result = analyze(code, "reader.go")
    assert any("deprecated" in s.lower() or "migrate" in s.lower() for s in result["suggestions"])


def test_suggestion_obsolete():
    """Obsolete patterns should produce an obsolete suggestion."""
    code = "// Copyright 2015 Company Inc.\npackage main\n"
    result = analyze(code, "main.go")
    assert any("obsolete" in s.lower() or "update" in s.lower() for s in result["suggestions"])
