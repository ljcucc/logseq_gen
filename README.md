# Logseq Generator

A universal generation tool for Logseq, creating Markdown files from structured assets.

## Description

This project provides a universal generation tool designed to create Markdown files in the `pages/` directory based on the content found in the `assets/` directory. It's specifically tailored to process `index.ini` and `index.md` files within a structured asset hierarchy.

**How it works:**

For every `index.ini` file found in `assets/xxx/yyy/zzz/index.ini`, the tool generates a corresponding Markdown file at `pages/xxx___yyy___zzz.md`.

**Example:**

Given the following structure and content in `assets/xxx/yyy/zzz/`:

`index.ini`:
```ini
[header]
content="index.md"

[properties]
property_a=10
property_b=20
property_c=30
```

`index.md`:
```markdown
- test a
- test b
    - test b.1
- test c
```

The tool will generate `/pages/xxx_yyy_zzz.md` with the following content:

```markdown
generated:: true
property_a:: 10
property_b:: 20
property_c:: 30

- test a
- test b
    - test b.1
- test c
```

**Important Notes:**
*   All generated pages will have `generated:: true` as the very first line of the file.
*   The tool includes `clear build` functionality to delete all Markdown files in `pages/` that contain `generated:: true` (identified by splitting the first line by `::` and trimming "generated" and "true").
*   The `build markdown` process will first perform a `clear build`.
*   The tool searches for all possible `index.ini` files under the `assets/` directory to generate corresponding pages.

## Features

*   Automated Markdown generation from structured asset files.
*   Support for `index.ini` for metadata and `index.md` for content.
*   Clear build functionality to remove previously generated files.
*   Designed for use with Logseq, leveraging its page property syntax.

## Installation

_Describe how users can install your project. For example, if it's a Python script, mention cloning the repository and any dependencies._

```bash
# Example:
git clone https://github.com/ljcucc/logseq_gen.git
cd logseq_gen
# If there are dependencies, list them here, e.g., pip install -r requirements.txt
```

## Usage

_Explain how to use the tool. Provide command-line examples if applicable._

```bash
# Example:
python generate.py build
python generate.py clear
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
