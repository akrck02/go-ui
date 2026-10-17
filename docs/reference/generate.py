#!/usr/bin/env python3
"""Writes the package reference in docs/reference from the Go doc comments.

Run from the repository's root after changing a public API:

    python3 docs/reference/generate.py

Each page is the package's comment, then its functions, constants and types, each type with its
declaration and the first sentence of every constructor's and method's comment. The doc comments
are the source of truth: improve them, not the generated pages.
"""
import os
import re
import subprocess
import sys

MODULE = "github.com/akrck02/go-ui"

# (folder, page, operating system for go doc, architecture, group)
PACKAGES = [
    (".", "ui.md", "", "", "core"),
    ("window", "window.md", "", "", "core"),
]

# Turns `go doc -all` of a package into a Markdown reference: functions, and each type with its
# declaration, constructors and methods, each with the first sentence of its comment.
import re, subprocess, sys

def first_sentence(text):
    text = " ".join(text.split())
    match = re.match(r"(.+?\.)(\s|$)", text)
    return (match.group(1) if match else text).replace("|", "\\|")

def parse(output):
    blocks, current = [], None
    section = None
    lines = output.split("\n")
    index = 0
    entries = []  # (kind, declaration lines, doc)
    while index < len(lines):
        line = lines[index]
        if line in ("CONSTANTS", "VARIABLES", "FUNCTIONS", "TYPES"):
            section = line; index += 1; continue
        if section and line and not line.startswith(" ") and not line.startswith("\t"):
            declaration = [line]
            index += 1
            if line.count("`") % 2 == 1:  # a raw string over many lines (a shader): skip to its end
                while index < len(lines) and lines[index].count("`") % 2 == 0:
                    index += 1
                index += 1
            if line.rstrip().endswith(("{", "(")):
                while index < len(lines) and not (lines[index].startswith(("}", ")"))):
                    declaration.append(lines[index]); index += 1
                if index < len(lines):
                    declaration.append(lines[index]); index += 1
            doc = []
            while index < len(lines) and (lines[index].startswith("    ") or lines[index] == ""):
                if lines[index].startswith("    "):
                    doc.append(lines[index].strip())
                index += 1
                if index < len(lines) and lines[index] == "" and index + 1 < len(lines) and not lines[index+1].startswith("    "):
                    index += 1; break
            entries.append((section, declaration, " ".join(doc)))
            continue
        index += 1
    return entries

def signature(line):
    return line.strip().rstrip("{").strip()

def render_output(output):
    header = output.split("\n")
    entries = parse(output)
    out = []
    comment = []
    for line in header[1:]:
        if line in ("CONSTANTS", "VARIABLES", "FUNCTIONS", "TYPES"):
            break
        comment.append(line)
    package_comment = "\n".join(comment).strip()
    if package_comment:
        out.append(package_comment + "\n")
    functions = [e for e in entries if e[0] == "FUNCTIONS"]
    constants = [e for e in entries if e[0] in ("CONSTANTS", "VARIABLES")]
    types = [e for e in entries if e[0] == "TYPES"]
    if functions:
        out.append("### Functions\n\n| Function | |\n|---|---|")
        for _, declaration, doc in functions:
            out.append(f"| `{signature(declaration[0])}` | {first_sentence(doc)} |")
        out.append("")
    if constants:
        out.append("### Constants and variables\n")
        for _, declaration, doc in constants:
            if "`" in declaration[0]:  # a shader's source: its name is enough
                declaration = [declaration[0].split("=")[0].rstrip() + " = …"]
            out.append("```go\n" + "\n".join(declaration) + "\n```\n")
            if doc: out.append(first_sentence(doc) + "\n")
    current = None
    rows = []
    def flush():
        if rows:
            out.append("| Function or method | |\n|---|---|")
            out.extend(rows)
            out.append("")
        rows.clear()
    for _, declaration, doc in types:
        head = declaration[0]
        if head.startswith("type "):
            flush()
            name = head.split()[1]
            out.append(f"### {name}\n")
            if doc: out.append(" ".join(doc.split()) .replace("|","\\|") + "\n")
            decl = "\n".join(declaration)
            out.append("```go\n" + decl + "\n```\n")
        elif head.startswith("func "):
            rows.append(f"| `{signature(head)}` | {first_sentence(doc)} |")
        else:
            out.append("```go\n" + "\n".join(declaration) + "\n```\n")
            if doc: out.append(first_sentence(doc) + "\n")
    flush()
    return "\n".join(out)



def write_page(folder, page, goos, goarch):
    environment = dict(os.environ)
    if goos:
        environment["GOOS"], environment["GOARCH"] = goos, goarch
    output = subprocess.run(["go", "doc", "-all", "./" + folder], capture_output=True, text=True,
                            check=True, env=environment).stdout
    body = render_output(output)
    import_path = MODULE + ("/" + folder if folder != "." else "")
    built = f"\n> Built only with `GOOS={goos}`.\n" if goos else ""
    text = (f"# {folder if folder != '.' else MODULE.split('/')[-1]}\n\n```go\nimport \"{import_path}\"\n```\n{built}\n"
            f"{body}\n---\n\n[← Reference](README.md) · generated from the doc comments by `generate.py`\n")
    with open(os.path.join("docs", "reference", page), "w") as file:
        file.write(text)


if __name__ == "__main__":
    for folder, page, goos, goarch, _ in PACKAGES:
        write_page(folder, page, goos, goarch)
        print("wrote", page)
