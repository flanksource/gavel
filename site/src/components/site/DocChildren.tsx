import { Link } from "react-router-dom";
import { listDocs } from "@/lib/mdx";

export default function DocChildren({ parent }: { parent: string }) {
  return <ul>
    {listDocs().filter((doc) => doc.frontmatter.parent === parent).map((doc) => <li key={doc.slug}>
      <Link to={`/docs/${doc.slug}`}>{doc.frontmatter.title}</Link>
      {doc.frontmatter.description && <span className="ml-2 text-sm text-muted-foreground">{doc.frontmatter.description}</span>}
    </li>)}
  </ul>;
}
