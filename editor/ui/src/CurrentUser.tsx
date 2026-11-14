import { useEffect, useState } from "react";
import { Avatar } from "@flanksource/clicky-ui/data";
import { request } from "./api";

/** The actor this editor authors comments as, from GET /api/me. */
export type Me = { kind: string; id: string; label?: string };

/** The top-bar avatar of whoever the editor's comments are written as. */
export function CurrentUser({ onError }: { onError: (message: string) => void }) {
  const [me, setMe] = useState<Me>();
  useEffect(() => {
    request<Me>("/api/me").then(setMe, (cause) => onError(`Loading the signed-in user: ${String(cause)}`));
  }, [onError]);
  if (!me) return null;
  const name = me.label || me.id;
  return <Avatar alt={name} colorKey={me.id} size="sm" title={`Commenting as ${name}`} />;
}
