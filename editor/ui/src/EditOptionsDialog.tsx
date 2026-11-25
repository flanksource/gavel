import { useEffect, useState } from "react";
import { Button, Combobox, Field, Modal } from "@flanksource/clicky-ui/components";
import { request, type EditListing, type ListingConfig, type OpenedFiles } from "./api";

/** The answer to PUT /api/edit/options: the listing now in effect and the files it lists. */
export type ListingApplied = OpenedFiles & { listing: EditListing };

const sameList = (a: string[] = [], b: string[] = []) => a.length === b.length && a.every((value, index) => value === b[index]);

/** Where the listing value in effect comes from, for the field's helper text. */
export function listingSource(listing: EditListing, key: keyof ListingConfig): string {
  const value = listing[key] ?? [];
  if (listing.flags[key]?.length && sameList(value, listing.flags[key])) return `--${key === "extensions" ? "ext" : "include-ignored"} flag`;
  if (listing.configured[key]?.length && sameList(value, listing.configured[key])) return `${listing.dir}/.gavel.yaml`;
  if (value.length === 0) return "default";
  return "applied in this editor, not saved";
}

/**
 * Chooses which files the open folders list: their extensions, and gitignored
 * patterns listed anyway. Apply rescans the folders; Save also writes the
 * edit section of the first folder's .gavel.yaml.
 */
export function EditOptionsDialog({ onApplied, onClose }: { onApplied: (applied: ListingApplied) => void; onClose: () => void }) {
  const [listing, setListing] = useState<EditListing>();
  const [extensions, setExtensions] = useState<string[]>([]);
  const [includeIgnored, setIncludeIgnored] = useState<string[]>([]);
  const [error, setError] = useState<string>();
  const [busy, setBusy] = useState(false);
  useEffect(() => {
    void request<EditListing>("/api/edit/options")
      .then((loaded) => {
        setListing(loaded);
        setExtensions(loaded.extensions ?? []);
        setIncludeIgnored(loaded.includeIgnored ?? []);
      })
      .catch((cause) => setError(String(cause)));
  }, []);

  const apply = async (save: boolean) => {
    setBusy(true);
    setError(undefined);
    try {
      onApplied(await request<ListingApplied>("/api/edit/options", { extensions, includeIgnored, save }, "PUT"));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause));
    } finally {
      setBusy(false);
    }
  };
  const extensionOptions = [...new Set([...(listing?.defaults ?? []), ...extensions])].map((ext) => ({ value: ext, label: ext }));

  return (
    <Modal
      open
      onClose={onClose}
      title="View options"
      subtitle={listing ? `Files listed in ${listing.dir}` : undefined}
      size="md"
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button variant="outline" onClick={() => void apply(false)} disabled={!listing || busy}>
            Apply
          </Button>
          <Button onClick={() => void apply(true)} disabled={!listing || busy}>
            {busy ? "Saving…" : "Save to .gavel.yaml"}
          </Button>
        </div>
      }
    >
      <div className="flex flex-col gap-4">
        <Field
          label="Extensions"
          htmlFor="edit-extensions"
          helper={
            listing &&
            `${extensions.length ? "Only these are listed" : "Empty lists the built-in documents and images"}; extensions without a richer view open as source. From: ${listingSource(listing, "extensions")}.`
          }
        >
          <Combobox
            id="edit-extensions"
            ariaLabel="Extensions"
            multiple
            variant="tags"
            allowCustomValue
            separators={[",", " "]}
            placeholder="Built-in documents and images"
            value={extensions}
            onChange={setExtensions}
            options={extensionOptions}
            onNew={(value) => {
              const ext = value.trim().startsWith(".") ? value.trim() : `.${value.trim()}`;
              return ext.length > 1 ? { value: ext.toLowerCase(), label: ext.toLowerCase() } : null;
            }}
          />
        </Field>
        <Field
          label="Show despite .gitignore"
          htmlFor="edit-include-ignored"
          helper={listing && `Folder-relative globs or directories, e.g. .tmp or dist/**/*.md. From: ${listingSource(listing, "includeIgnored")}.`}
        >
          <Combobox
            id="edit-include-ignored"
            ariaLabel="Show despite .gitignore"
            multiple
            variant="tags"
            allowCustomValue
            separators={[","]}
            placeholder="None"
            value={includeIgnored}
            onChange={setIncludeIgnored}
            options={includeIgnored.map((pattern) => ({ value: pattern, label: pattern }))}
            onNew={(value) => (value.trim() ? { value: value.trim(), label: value.trim() } : null)}
          />
        </Field>
        {error && (
          <p role="alert" className="text-sm text-destructive">
            {error}
          </p>
        )}
      </div>
    </Modal>
  );
}
