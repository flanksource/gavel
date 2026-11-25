/** A new-file template from GET /api/templates. */
export type Template = {
  id: string;
  group: string;
  label: string;
  description: string;
  /** The suggested file name. */
  name: string;
  /** The ending every file created from the template must have. */
  suffix: string;
  source: string;
};

/**
 * The file name kept when the chosen template changes: a name the user typed keeps
 * its stem under the new template's suffix; the previous suggestion becomes the new one.
 */
export function retargetName(name: string, from: Template | undefined, to: Template): string {
  if (!from || name === from.name || !name.toLowerCase().endsWith(from.suffix)) return to.name;
  return name.slice(0, name.length - from.suffix.length) + to.suffix;
}

/** `name` under `folder`, both relative to the tree root; `folder` "" is the root. */
export const joinPath = (folder: string, name: string) => (folder ? `${folder.replace(/\/+$/, "")}/${name}` : name);

/** The templates grouped in catalog order. */
export function templateGroups(templates: readonly Template[]): [string, Template[]][] {
  const groups = new Map<string, Template[]>();
  for (const template of templates) groups.set(template.group, [...(groups.get(template.group) ?? []), template]);
  return [...groups];
}
