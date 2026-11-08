import { useEffect, useRef, useState } from "react";
import { cn } from "@/lib/utils";

interface TypingAnimationProps {
  text: string;
  className?: string;
  speedMs?: number;
}

export default function TypingAnimation({ text, className, speedMs = 40 }: TypingAnimationProps) {
  const [shown, setShown] = useState("");
  const [started, setStarted] = useState(false);
  const ref = useRef<HTMLSpanElement>(null);

  useEffect(() => {
    const node = ref.current;
    if (!node) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) {
          setStarted(true);
          observer.disconnect();
        }
      },
      { threshold: 0.4 },
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    if (!started) return;
    let i = 0;
    const id = window.setInterval(() => {
      i += 1;
      setShown(text.slice(0, i));
      if (i >= text.length) window.clearInterval(id);
    }, speedMs);
    return () => window.clearInterval(id);
  }, [started, text, speedMs]);

  return (
    <span ref={ref} className={cn("font-mono", className)}>
      {shown}
      <span className="ml-0.5 inline-block h-[1em] w-[0.5ch] animate-pulse bg-current align-text-bottom" aria-hidden />
    </span>
  );
}
