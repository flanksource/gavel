import sys
from pathlib import Path

import pymupdf


source, destination = sys.argv[1:]
with pymupdf.open(source) as document:
    markdown = document.to_markdown(write_images=True, image_path="media", filename="document", show_progress=False)
if not isinstance(markdown, str) or not markdown.strip():
    raise ValueError(f"No Markdown content extracted from {source}")
Path(destination).write_text(markdown, encoding="utf-8")
