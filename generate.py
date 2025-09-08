
import os
import sys
import configparser

# --- Constants ---
ROOT_DIR = os.path.dirname(os.path.abspath(__file__))
ASSETS_DIR = os.path.join(ROOT_DIR, "assets")
PAGES_DIR = os.path.join(ROOT_DIR, "pages")
GENERATED_MARKER = "generated:: true"

# --- Core Functions ---

def clear_build():
    """
    Removes all Markdown files in the pages directory that are marked as generated.
    """
    if not os.path.exists(PAGES_DIR):
        print("Pages directory does not exist. Nothing to clear.")
        return

    print(f"Clearing generated files from {PAGES_DIR}...")
    for filename in os.listdir(PAGES_DIR):
        if filename.endswith(".md"):
            file_path = os.path.join(PAGES_DIR, filename)
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    first_line = f.readline().strip()
                    parts = [part.strip() for part in first_line.split("::", 1)]
                    if len(parts) == 2 and parts[0] == "generated" and parts[1] == "true":
                        os.remove(file_path)
                        print(f"Removed {filename}")
            except Exception as e:
                print(f"Error processing file {filename}: {e}")
    print("Clear build finished.")

def build_markdown():
    """
    Generates Markdown files in the pages directory based on index.ini files
    found in the assets directory.
    """
    # 1. Clear previous build
    clear_build()

    # 2. Ensure pages directory exists
    os.makedirs(PAGES_DIR, exist_ok=True)

    print(f"\nStarting build process from {ASSETS_DIR}...")
    # 3. Find all index.ini files and generate pages
    for root, _, files in os.walk(ASSETS_DIR):
        if "index.ini" in files:
            ini_path = os.path.join(root, "index.ini")
            print(f"Found index.ini at: {ini_path}")
            
            try:
                # Parse INI file
                config = configparser.ConfigParser()
                config.read(ini_path, encoding='utf-8')

                # Get content file path from [header]
                if not config.has_section("header") or not config.has_option("header", "content"):
                    print(f"Skipping {ini_path}: Missing [header] section or 'content' option.")
                    continue
                
                content_filename = config.get("header", "content").strip('"')
                content_filepath = os.path.join(root, content_filename)

                if not os.path.exists(content_filepath):
                    print(f"Skipping {ini_path}: Content file '{content_filepath}' not found.")
                    continue

                # Read content from the specified markdown file
                with open(content_filepath, 'r', encoding='utf-8') as f:
                    md_content = f.read()

                # Build the output content
                output_content = [GENERATED_MARKER]
                
                # Add properties from [properties]
                if config.has_section("properties"):
                    for key, value in config.items("properties"):
                        output_content.append(f"{key}:: {value}")
                
                # Combine properties and markdown content
                final_content = "\n".join(output_content) + "\n\n" + md_content

                # Determine the output filename
                relative_path = os.path.relpath(root, ASSETS_DIR)
                if relative_path == ".":
                    # Handle case where index.ini is directly in assets/
                    output_filename_base = "index"
                else:
                    output_filename_base = relative_path.replace(os.path.sep, "___")
                
                output_filename = f"{output_filename_base}.md"
                output_filepath = os.path.join(PAGES_DIR, output_filename)

                # Write the generated file
                with open(output_filepath, 'w', encoding='utf-8') as f:
                    f.write(final_content)
                print(f"Generated {output_filepath}")

            except Exception as e:
                print(f"Error processing {ini_path}: {e}")
    
    print("Build process finished.")


# --- Main Execution ---

if __name__ == "__main__":
    if len(sys.argv) > 1:
        command = sys.argv[1].lower()
        if command == "build":
            build_markdown()
        elif command == "clear":
            clear_build()
        else:
            print(f"Unknown command: {command}")
            print("Usage: python generate.py [build|clear]")
    else:
        print("Usage: python generate.py [build|clear]")
        print("Defaulting to 'build'.")
        build_markdown()
