import { Link } from "react-router-dom";

export default function NotFound() {
  return (
    <div className="mx-auto max-w-2xl px-6 py-24 text-center">
      <h1 className="text-5xl font-bold">404</h1>
      <p className="mt-4 text-muted-foreground">That page is not here.</p>
      <Link to="/" className="mt-8 inline-block text-primary hover:underline">
        ← Back to home
      </Link>
    </div>
  );
}
