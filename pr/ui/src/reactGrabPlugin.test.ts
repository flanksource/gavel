import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import pluginSource from '../assets/react-grab-plugin.js?raw';

type Plugin = {
  hooks: { transformCopyContent: (content: string, elements: Element[]) => Promise<string> };
  actions: { id: string; onAction: (ctx: Record<string, unknown>) => void }[];
};

const COMPONENT = 'BadgeTrigger';
const STACK = 'in BadgeTrigger (at math-badges.tsx:58)';

function loadPlugin(): Plugin {
  let registered: Plugin | undefined;
  Object.assign(window, {
    __REACT_GRAB__: {
      registerPlugin: (plugin: Plugin) => { registered = plugin; },
      setToolbarState: vi.fn(),
      getDisplayName: () => COMPONENT,
      getSource: () => Promise.resolve({ filePath: 'math-badges.tsx', lineNumber: 58 }),
      getStackContext: () => Promise.resolve(STACK),
    },
  });
  new Function(pluginSource.replaceAll('__GAVEL_ORIGIN__', 'http://gavel.test'))();
  if (!registered) throw new Error('plugin did not register with React Grab');
  return registered;
}

function renderRows() {
  document.body.innerHTML = `
    <main>
      <ul>
        <li data-result-row="Gross"><button data-badge="expr" aria-label="EXPR of Gross">EXPR</button></li>
        <li data-result-row="Net"><button data-badge="expr" aria-label="EXPR of Net">EXPR</button></li>
      </ul>
    </main>`;
  return document.querySelectorAll('button')[1]!;
}

describe('gavel React Grab plugin', () => {
  beforeEach(() => {
    HTMLDialogElement.prototype.showModal = vi.fn();
  });
  afterEach(() => {
    Reflect.deleteProperty(window, '__REACT_GRAB__');
    document.body.innerHTML = '';
  });

  it('anchors the copied body with a selector that resolves back to the grabbed element', async () => {
    const plugin = loadPlugin();
    const grabbed = renderRows();

    const body = await plugin.hooks.transformCopyContent('', [grabbed]);

    const selector = /\*\*Selector:\*\* `([^`]+)`/.exec(body)?.[1];
    expect(selector).toBeDefined();
    expect(document.querySelector(selector!)).toBe(grabbed);
    expect(body).toContain(STACK);
  });

  it('titles the todo with the component and the element accessible name', async () => {
    const plugin = loadPlugin();
    const grabbed = renderRows();
    const todo = plugin.actions.find((action) => action.id === 'gavel-todo')!;

    todo.onAction({ element: grabbed, componentName: COMPONENT, tagName: 'button' });
    await vi.waitFor(() => expect(document.querySelector('iframe.gavel-rg-frame')).not.toBeNull());

    const src = new URL(document.querySelector<HTMLIFrameElement>('iframe.gavel-rg-frame')!.src);
    expect(src.searchParams.get('title')).toBe(`${COMPONENT}: EXPR of Net`);
  });
});
