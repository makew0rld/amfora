#!/usr/bin/env python3

# Formatted with black.

import shutil
import subprocess
import sys
import os
import md2gemini

TMP_WIKI_CLONE = "/tmp/amfora.wiki"


def md2gem(markdown):
    return md2gemini.md2gemini(
        markdown,
        links="copy",
        plain=False,
        strip_html=True,
        md_links=True,
        link_func=link_func,
    )


def link_func(link):
    if "://" in link:
        # Absolute URL
        return link

    # Link to other wiki page
    return link + ".gmi"


def run_cmd(*args):
    proc = subprocess.run(args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    if proc.returncode != 0:
        print(
            "Command "
            + " ".join(args)
            + "failed with exit code "
            + str(proc.returncode)
        )
        print("Output was:")
        print()
        print(proc.stdout.decode())
        sys.exit(1)


# Delete leftover git repo
try:
    shutil.rmtree(TMP_WIKI_CLONE)
except FileNotFoundError:
    pass

os.mkdir(TMP_WIKI_CLONE)

run_cmd(
    "git",
    "clone",
    "--depth",
    "1",
    "https://github.com/makeworld-the-better-one/amfora.wiki.git",
    TMP_WIKI_CLONE,
)

# Save special files

with open(os.path.join(TMP_WIKI_CLONE, "_Footer.md"), "r") as f:
    footer = md2gem(f.read())

# Get files
(_, _, files) = next(os.walk(TMP_WIKI_CLONE))

# Create list of pages
pages = "## Pages\n\n=>.. Home\n"
for file in files:

    if file in ["_Footer.md", "_Sidebar.md", "Home.md"]:
        continue
    if not file.endswith(".md"):
        continue
    pages += "=>" + file[:-2] + "gmi " + file[:-3].replace("-", " ") + "\n"

pages += "\n\n"

for file in files:
    filepath = os.path.join(TMP_WIKI_CLONE, file)

    if file in ["_Footer.md", "_Sidebar.md"]:
        continue
    if not file.endswith(".md"):
        # Could be a resource like an image file, copy it
        shutil.copyfile(filepath, file)
        continue

    # Markdown file

    with open(filepath, "r") as f:
        gemtext = md2gem(f.read())

    # Add title, sidebar, footer
    gemtext = "# " + file[:-3].replace("-", " ") + "\n\n" + pages + gemtext
    gemtext += "\n\n\n\n" + footer

    if file == "Home.md":
        file = "index.md"

    new_name = file[:-2] + "gmi"

    with open(new_name, "w") as f:
        f.write(gemtext)