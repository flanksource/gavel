export default function SiteFooter() {
  return (
    <footer className="border-t bg-background">
      <div className="mx-auto flex max-w-6xl flex-col gap-2 px-6 py-8 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
        <p>
          <a
            href="https://github.com/flanksource/gavel"
            target="_blank"
            rel="noreferrer"
            className="hover:text-foreground"
          >
            github.com/flanksource/gavel
          </a>
        </p>
        <p>© Flanksource</p>
      </div>
    </footer>
  );
}
