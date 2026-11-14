import sys
from io import BytesIO
from pathlib import Path, PurePosixPath

from docx2md import Converter, DocxFile, DocxMedia
from PIL import Image


source, destination = sys.argv[1:]
document = DocxFile(source)
try:
    media = DocxMedia(document)
    target = Path(destination).parent
    for item in media.hash.values():
        relative = PurePosixPath(item.path)
        if relative.is_absolute() or ".." in relative.parts or "\\" in item.path:
            raise ValueError(f"Unsafe DOCX media path: {item.path}")
        data = document.read("word/" + item.path)
        if not data:
            raise ValueError(f"Missing or empty DOCX media: {item.path}")
        output = target / item.path
        output.parent.mkdir(parents=True, exist_ok=True)
        if output.suffix.lower() in {".emf", ".wmf"}:
            output.write_bytes(data)
            item.alt_path = item.path + ".svg"
        elif item.use_alt:
            with Image.open(BytesIO(data)) as image:
                image.save(target / item.alt_path, format="PNG")
        else:
            output.write_bytes(data)
    content = Converter(document.document(), media, True).convert()
    if not content.strip():
        raise ValueError(f"Empty Markdown extracted from {source}")
    Path(destination).write_text(content, encoding="utf-8")
finally:
    document.close()
