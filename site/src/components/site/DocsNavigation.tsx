import { Link } from "react-router-dom";
import type { DocNode } from "@/lib/docs";
import { cn } from "@/lib/utils";

function containsActive(node: DocNode, active?: string): boolean {
  return node.doc.slug === active || node.children.some((child) => containsActive(child, active));
}

export default function DocsNavigation({ nodes, active }: { nodes: DocNode[]; active?: string }) {
  return (
    <ul className="space-y-1 text-sm">
      {nodes.map((node) => {
        const link = <Link to={`/docs/${node.doc.slug}`} aria-current={active === node.doc.slug ? "page" : undefined} className={cn(
          "rounded px-2 py-1.5 transition-colors",
          active === node.doc.slug ? "bg-muted font-medium text-foreground" : "text-muted-foreground hover:text-foreground",
        )}>{node.doc.frontmatter.title}</Link>;
        return <li key={node.doc.slug}>
          {node.children.length ? <details open={containsActive(node, active)}>
            <summary className="cursor-pointer py-1.5">{link}</summary>
            <div className="ml-3 border-l border-border pl-2">
              <DocsNavigation nodes={node.children} active={active} />
            </div>
          </details> : <div className="py-1.5">{link}</div>}
        </li>;
      })}
    </ul>
  );
}
