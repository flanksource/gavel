import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

interface AnimatedGradientTextProps {
  children: ReactNode;
  className?: string;
}

export default function AnimatedGradientText({ children, className }: AnimatedGradientTextProps) {
  return (
    <span
      className={cn(
        "inline-block bg-[linear-gradient(to_right,theme(colors.sky.400),theme(colors.blue.500),theme(colors.indigo.400),theme(colors.sky.400))] bg-[length:200%_auto] bg-clip-text text-transparent animate-[gradient_4s_linear_infinite]",
        className,
      )}
      style={{
        WebkitBackgroundClip: "text",
      }}
    >
      {children}
    </span>
  );
}
