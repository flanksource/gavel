import { Link } from "react-router-dom";
import ThemeToggle from "./ThemeToggle";

export default function SiteHeader() {
  return (
    <header className="sticky top-0 z-40 border-b bg-background/80 backdrop-blur">
      <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-6">
        <Link to="/" className="flex items-center gap-2 font-mono text-lg font-semibold">
          <img src="/favicon.svg" alt="" className="h-6 w-6" width={24} height={24} />
          <span>gavel</span>
        </Link>
        <nav className="flex items-center gap-5 text-sm text-muted-foreground">
          <Link to="/docs" className="hover:text-foreground">Docs</Link>
          <Link to="/blog" className="hover:text-foreground">Blog</Link>
          <a
            href="https://github.com/flanksource/gavel"
            className="hover:text-foreground"
            target="_blank"
            rel="noreferrer"
          >
            GitHub
          </a>
          <ThemeToggle />
        </nav>
      </div>
    </header>
  );
}
