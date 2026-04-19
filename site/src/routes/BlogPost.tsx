import { useParams } from "react-router-dom";

export default function BlogPost() {
  const { slug } = useParams();
  return (
    <div className="mx-auto max-w-3xl px-6 py-16">
      <h1 className="text-3xl font-bold">{slug}</h1>
      <p className="mt-2 text-muted-foreground">MDX post rendering lands in Phase 4.</p>
    </div>
  );
}
