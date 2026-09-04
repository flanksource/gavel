import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import mdx from "@mdx-js/rollup";
import tailwindcss from "@tailwindcss/vite";
import remarkGfm from "remark-gfm";
import remarkFrontmatter from "remark-frontmatter";
import remarkMdxFrontmatter from "remark-mdx-frontmatter";
import rehypePrettyCode from "rehype-pretty-code";
import path from "node:path";
import type {
  CodeToHastOptions,
  DecorationItem,
  ShikiTransformer,
} from "shiki";

type CodeAnnotation =
  | { kind: "line"; classes: string[]; count: number; note?: string }
  | { kind: "word"; value: string; count: number };

type PendingAnnotation = CodeAnnotation & { remaining: number };

type LineDecoration = {
  classes: Set<string>;
  note?: string;
};

const codeAnnotationPattern = /\[!code\s+([^\]]+)\]/g;
const codeAnnotationCommentPattern =
  /(\s*(?:\/\/|#|<!--)\s*)?\[!code\s+[^\]]+\](?:\s*-->)?/g;

function parseCount(value: string | undefined, fallback = 1) {
  if (!value) {
    return fallback;
  }
  const count = Number.parseInt(value, 10);
  return Number.isFinite(count) && count > 0 ? count : fallback;
}

function parseWordAnnotation(value: string) {
  const countMatch = value.match(/:(\d+)$/);
  if (!countMatch?.[1]) {
    return { value, count: Number.POSITIVE_INFINITY };
  }
  return {
    value: value.slice(0, -countMatch[0].length),
    count: parseCount(countMatch[1]),
  };
}

function parseCodeAnnotation(raw: string): CodeAnnotation | null {
  const [directive, ...rest] = raw.trim().split(":");
  const value = rest.join(":").trim();

  switch (directive) {
    case "++":
      return { kind: "line", classes: ["diff", "add"], count: 1 };
    case "--":
      return { kind: "line", classes: ["diff", "remove"], count: 1 };
    case "focus":
      return { kind: "line", classes: ["focused"], count: parseCount(value) };
    case "highlight":
      return {
        kind: "line",
        classes: ["highlighted"],
        count: parseCount(value),
      };
    case "info":
    case "warning":
    case "error":
      return {
        kind: "line",
        classes: ["highlighted", directive],
        count: parseCount(value),
      };
    case "note":
      return {
        kind: "line",
        classes: ["annotated"],
        count: 1,
        note: value,
      };
    case "word": {
      const annotation = parseWordAnnotation(value);
      return {
        kind: "word",
        value: annotation.value,
        count: annotation.count,
      };
    }
    default:
      return null;
  }
}

function addLineDecoration(
  map: Map<number, LineDecoration>,
  line: number,
  annotation: Extract<CodeAnnotation, { kind: "line" }>,
) {
  const decoration = map.get(line) ?? { classes: new Set<string>() };
  annotation.classes.forEach((className) => decoration.classes.add(className));
  if (annotation.note) {
    decoration.note = annotation.note;
  }
  map.set(line, decoration);
}

function addWordDecorations(
  decorations: DecorationItem[],
  line: number,
  text: string,
  value: string,
) {
  if (!value) {
    return;
  }

  let start = 0;
  while (start < text.length) {
    const index = text.indexOf(value, start);
    if (index === -1) {
      break;
    }
    decorations.push({
      start: { line, character: index },
      end: { line, character: index + value.length },
      properties: { class: "highlighted-word" },
      alwaysWrap: true,
    });
    start = index + value.length;
  }
}

function collectDecoratedCode(code: string) {
  const pending: PendingAnnotation[] = [];
  const lineDecorations = new Map<number, LineDecoration>();
  const wordDecorations: DecorationItem[] = [];
  const cleanedLines: string[] = [];

  for (const sourceLine of code.split("\n")) {
    const annotations = [...sourceLine.matchAll(codeAnnotationPattern)]
      .map((match) => parseCodeAnnotation(match[1] ?? ""))
      .filter((annotation): annotation is CodeAnnotation => annotation != null);
    const cleanedLine = sourceLine
      .replace(codeAnnotationCommentPattern, "")
      .replace(/[ \t]+$/, "");

    if (annotations.length > 0 && cleanedLine.trim() === "") {
      pending.push(
        ...annotations.map((annotation) => ({
          ...annotation,
          remaining: annotation.count,
        })),
      );
      continue;
    }

    const line = cleanedLines.length;
    cleanedLines.push(cleanedLine);

    for (const annotation of pending) {
      if (annotation.kind === "line") {
        addLineDecoration(lineDecorations, line, annotation);
      } else {
        addWordDecorations(wordDecorations, line, cleanedLine, annotation.value);
      }
      annotation.remaining -= 1;
    }

    for (const annotation of annotations) {
      if (annotation.kind === "line") {
        addLineDecoration(lineDecorations, line, annotation);
      } else {
        addWordDecorations(wordDecorations, line, cleanedLine, annotation.value);
      }
    }

    for (let index = pending.length - 1; index >= 0; index--) {
      if (pending[index].remaining <= 0) {
        pending.splice(index, 1);
      }
    }
  }

  const decorations: DecorationItem[] = wordDecorations.filter(
    (decoration) =>
      typeof decoration.start !== "object" ||
      !lineDecorations.has(decoration.start.line),
  );
  for (const [line, decoration] of lineDecorations) {
    const text = cleanedLines[line] ?? "";
    if (text.length === 0) {
      continue;
    }
    decorations.push({
      start: { line, character: 0 },
      end: { line, character: text.length },
      properties: {
        class: [...decoration.classes].join(" "),
        ...(decoration.note ? { "data-annotation": decoration.note } : {}),
      },
    });
  }

  return {
    code: cleanedLines.join("\n"),
    decorations,
    hasAnnotations: decorations.length > 0,
    hasFocus: [...lineDecorations.values()].some((decoration) =>
      decoration.classes.has("focused"),
    ),
  };
}

function appendClass(
  properties: Record<string, unknown>,
  className: string | string[],
) {
  const classes = Array.isArray(className) ? className : [className];
  const current = properties.className ?? properties.class ?? [];
  const currentClasses = Array.isArray(current)
    ? current
    : String(current).split(/\s+/).filter(Boolean);

  properties.className = [...new Set([...currentClasses, ...classes])];
}

function codeAnnotationTransformer(): ShikiTransformer {
  const states = new WeakMap<
    object,
    { hasAnnotations: boolean; hasFocus: boolean }
  >();

  return {
    name: "gavel:code-annotations",
    preprocess(code: string, options: CodeToHastOptions) {
      const annotated = collectDecoratedCode(code);
      if (!annotated.hasAnnotations) {
        return code;
      }

      options.decorations ||= [];
      options.decorations.push(...annotated.decorations);
      states.set(this.meta, {
        hasAnnotations: annotated.hasAnnotations,
        hasFocus: annotated.hasFocus,
      });
      return annotated.code;
    },
    pre(element) {
      const state = states.get(this.meta);
      if (!state?.hasAnnotations) {
        return;
      }

      appendClass(element.properties, [
        "has-code-annotations",
        ...(state.hasFocus ? ["has-focused"] : []),
      ]);
    },
  };
}

const prettyCodeOptions = {
  theme: { light: "github-light", dark: "github-dark" },
  keepBackground: false,
  bypassInlineCode: true,
  defaultLang: {
    block: "plaintext",
  },
  transformers: [codeAnnotationTransformer()],
  onVisitLine(element: any) {
    if (element.children.length === 0) {
      element.children = [{ type: "text", value: " " }];
    }
  },
  onVisitHighlightedLine(element: any) {
    element.properties.className = [
      ...(element.properties.className || []),
      "highlighted-line",
    ];
  },
  onVisitHighlightedChars(element: any) {
    element.properties.className = [
      ...(element.properties.className || []),
      "highlighted-chars",
    ];
  },
  onVisitTitle(element: any) {
    element.properties.className = [
      ...(element.properties.className || []),
      "code-title",
    ];
  },
  onVisitCaption(element: any) {
    element.properties.className = [
      ...(element.properties.className || []),
      "code-caption",
    ];
  },
};

export default defineConfig({
  plugins: [
    {
      enforce: "pre",
      ...mdx({
        remarkPlugins: [
          remarkGfm,
          remarkFrontmatter,
          [remarkMdxFrontmatter, { name: "frontmatter" }],
        ],
        rehypePlugins: [
          [rehypePrettyCode, prettyCodeOptions],
        ],
        providerImportSource: "@mdx-js/react",
      }),
    },
    react(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
});
