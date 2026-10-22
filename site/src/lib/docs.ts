import type { MdxEntry } from "./mdx";

export interface DocNode {
  doc: MdxEntry;
  children: DocNode[];
}

export function buildDocTree(docs: MdxEntry[]): DocNode[] {
  const nodes = new Map<string, DocNode>();
  for (const doc of docs) {
    if (nodes.has(doc.slug)) throw new Error(`duplicate documentation slug: ${doc.slug}`);
    nodes.set(doc.slug, { doc, children: [] });
  }
  const roots: DocNode[] = [];
  for (const node of nodes.values()) {
    const ancestors = new Set([node.doc.slug]);
    let parent = node.doc.frontmatter.parent;
    while (parent) {
      if (ancestors.has(parent)) throw new Error(`documentation cycle at ${parent}`);
      ancestors.add(parent);
      const ancestor = nodes.get(parent);
      if (!ancestor) throw new Error(`missing parent ${parent} for ${node.doc.slug}`);
      parent = ancestor.doc.frontmatter.parent;
    }
    if (node.doc.frontmatter.parent) nodes.get(node.doc.frontmatter.parent)!.children.push(node);
    else roots.push(node);
  }
  return roots;
}
