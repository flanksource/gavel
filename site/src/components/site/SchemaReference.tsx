interface SchemaNode {
  type?: string | string[];
  title?: string;
  description?: string;
  $ref?: string;
  $defs?: Record<string, SchemaNode>;
  definitions?: Record<string, SchemaNode>;
  properties?: Record<string, SchemaNode>;
  required?: string[];
  items?: SchemaNode;
  additionalProperties?: boolean | SchemaNode;
  oneOf?: SchemaNode[];
  anyOf?: SchemaNode[];
  allOf?: SchemaNode[];
  enum?: unknown[];
  const?: unknown;
  default?: unknown;
  format?: string;
  minimum?: number;
  maximum?: number;
  minItems?: number;
  maxItems?: number;
  minLength?: number;
  maxLength?: number;
  pattern?: string;
  deprecated?: boolean;
}

interface Field {
  path: string;
  schema: SchemaNode;
  required: boolean;
}

function fields(schema: SchemaNode, path = ""): Field[] {
  const result: Field[] = [];
  for (const [name, child] of Object.entries(schema.properties ?? {})) {
    const childPath = path ? `${path}.${name}` : name;
    result.push({ path: childPath, schema: child, required: schema.required?.includes(name) ?? false });
  }
  if (schema.items) result.push({ path: `${path}[]`, schema: schema.items, required: false });
  if (typeof schema.additionalProperties === "object") {
    const childPath = `${path}.<name>`;
    result.push({ path: childPath, schema: schema.additionalProperties, required: false });
  }
  for (const [index, variant] of (schema.oneOf ?? schema.anyOf ?? schema.allOf ?? []).entries()) {
    result.push({ path: `${path} (${variant.title ?? `option ${index + 1}`})`, schema: variant, required: false });
  }
  return result;
}

function typeNames(schema: SchemaNode): string[] {
  if (schema.$ref) return [schema.$ref.split("/").pop()!];
  if (schema.const !== undefined) return [JSON.stringify(schema.const)];
  if (schema.oneOf || schema.anyOf || schema.allOf) {
    return [...new Set((schema.oneOf ?? schema.anyOf ?? schema.allOf ?? []).flatMap(typeNames))];
  }
  if (schema.type === "array" && schema.items) return typeNames(schema.items).map((type) => `${type}[]`);
  return Array.isArray(schema.type) ? schema.type : [schema.type ?? "value"];
}

function references(schema: SchemaNode): string[] {
  if (schema.$ref) return [schema.$ref.split("/").pop()!];
  if (schema.items) return references(schema.items);
  return (schema.oneOf ?? schema.anyOf ?? schema.allOf ?? []).flatMap(references);
}

function constraints(schema: SchemaNode): string[] {
  return Object.entries(schema)
    .filter(([key, value]) => ["enum", "default", "format", "minimum", "maximum", "minItems", "maxItems", "minLength", "maxLength", "pattern", "deprecated", "additionalProperties"].includes(key) && (key !== "additionalProperties" || typeof value === "boolean"))
    .map(([key, value]) => `${key}: ${typeof value === "string" ? value : JSON.stringify(value)}`);
}

function SchemaType({ schema, id }: { schema: SchemaNode; id: string }) {
  return <span className="ml-2 inline-flex flex-wrap gap-2 text-xs text-muted-foreground">
    <span>{typeNames(schema).join(" | ")}</span>
    {[...new Set(references(schema))].map((name) => <a key={name} href={`#${id}-${name}`} onClick={() => {
      const target = document.getElementById(`${id}-${name}`);
      if (target instanceof HTMLDetailsElement) target.open = true;
    }}>{name}</a>)}
  </span>;
}

function SchemaDescription({ schema }: { schema: SchemaNode }) {
  return <div className="mt-2 space-y-1 break-words text-sm">
    {schema.description && <div className="whitespace-pre-line">{schema.description}</div>}
    {constraints(schema).map((constraint) => <div key={constraint} className="font-mono text-xs text-muted-foreground">{constraint}</div>)}
  </div>;
}

function SchemaFields({ schema, id, path = "" }: { schema: SchemaNode; id: string; path?: string }) {
  return <div className="space-y-3">
    {fields(schema, path).map((field) => {
      const heading = <>
        <code className="break-all">{field.path}</code><SchemaType schema={field.schema} id={id} />
        {field.required && <span className="ml-2 text-xs text-muted-foreground">required</span>}
      </>;
      return fields(field.schema, field.path).length > 0
        ? <details key={field.path} data-path={field.path} className="rounded border border-border p-3">
          <summary className="cursor-pointer">{heading}</summary>
          <SchemaDescription schema={field.schema} />
          <div className="mt-3 border-l border-border pl-3"><SchemaFields schema={field.schema} id={id} path={field.path} /></div>
        </details>
        : <div key={field.path} className="rounded border border-border p-3">
          {heading}<SchemaDescription schema={field.schema} />
        </div>;
    })}
  </div>;
}

export default function SchemaReference({ schema, id }: { schema: unknown; id: string }) {
  const node = schema as SchemaNode;
  if (!node || typeof node !== "object") throw new Error(`Missing schema for ${id}`);
  return (
    <div id={id} className="scroll-mt-24">
      <SchemaFields schema={node} id={id} />
      {Object.keys(node.$defs ?? node.definitions ?? {}).length > 0 && <h3>Shared types</h3>}
      {Object.entries(node.$defs ?? node.definitions ?? {}).map(([name, definition]) => (
        <details key={name} id={`${id}-${name}`} className="my-3 scroll-mt-24 rounded border border-border p-3">
          <summary className="cursor-pointer font-semibold">{name}</summary>
          <SchemaType schema={definition} id={id} />
          <SchemaDescription schema={definition} />
          <div className="mt-3"><SchemaFields schema={definition} id={id} path={name} /></div>
        </details>
      ))}
    </div>
  );
}
