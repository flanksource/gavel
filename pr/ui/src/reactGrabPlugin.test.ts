import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import pluginSource from '../assets/react-grab-plugin.js?raw';

type Plugin = {
  hooks: { transformCopyContent: (content: string, elements: Element[]) => Promise<string> };
  actions: { id: string; onAction: (ctx: Record<string, unknown>) => void }[];
  setup?: (api: {
    activate: () => void;
    deactivate: () => void;
    isActive: () => boolean;
    isEnabled: () => boolean;
  }) => {
    hooks?: {
      onDeactivate?: () => void;
      onElementSelect?: (element: Element) => boolean | void;
    };
    cleanup?: () => void;
  };
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
  let shortcutCleanup: (() => void) | undefined;

  beforeEach(() => {
    HTMLDialogElement.prototype.showModal = vi.fn();
  });
  afterEach(() => {
    shortcutCleanup?.();
    shortcutCleanup = undefined;
    Reflect.deleteProperty(window, '__REACT_GRAB__');
    document.body.innerHTML = '';
  });

  it('opens one todo for the first selected element after Command+G', async () => {
    const plugin = loadPlugin();
    const grabbed = renderRows();
    const activate = vi.fn();
    const config = plugin.setup?.({
      activate,
      deactivate: vi.fn(),
      isActive: () => false,
      isEnabled: () => true,
    });
    shortcutCleanup = config?.cleanup;

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'g', metaKey: true, bubbles: true, cancelable: true }));
    expect(activate).toHaveBeenCalledOnce();
    expect(config?.hooks?.onElementSelect?.(grabbed)).toBe(true);
    expect(config?.hooks?.onElementSelect?.(document.querySelector('button')!)).toBe(true);
    await vi.waitFor(() => expect(document.querySelectorAll('iframe.gavel-rg-frame')).toHaveLength(1));

    const src = new URL(document.querySelector<HTMLIFrameElement>('iframe.gavel-rg-frame')!.src);
    expect(src.searchParams.get('title')).toBe(`${COMPONENT}: EXPR of Net`);
    expect(src.searchParams.get('body')).toContain(STACK);
  });

  it('toggles Command+G without claiming modifier variants, repeats, or editable fields', () => {
    const plugin = loadPlugin();
    let active = false;
    const activate = vi.fn(() => { active = true; });
    const deactivate = vi.fn(() => { active = false; });
    const config = plugin.setup?.({ activate, deactivate, isActive: () => active, isEnabled: () => true });
    shortcutCleanup = config?.cleanup;
    const input = document.createElement('input');
    document.body.appendChild(input);

    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'g', metaKey: true, bubbles: true }));
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'g', ctrlKey: true }));
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'g', metaKey: true, shiftKey: true }));
    expect(activate).not.toHaveBeenCalled();

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'g', metaKey: true }));
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'g', metaKey: true, repeat: true }));
    expect(activate).toHaveBeenCalledOnce();
    expect(deactivate).not.toHaveBeenCalled();

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'g', metaKey: true }));
    expect(deactivate).toHaveBeenCalledOnce();
    expect(config?.hooks?.onElementSelect?.(renderRows())).toBeUndefined();
  });

  it('removes the old shortcut listener when React Grab registers the plugin again', () => {
    const plugin = loadPlugin();
    const oldActivate = vi.fn();
    const oldDeactivate = vi.fn();
    const previous = plugin.setup?.({
      activate: oldActivate,
      deactivate: oldDeactivate,
      isActive: () => false,
      isEnabled: () => true,
    });
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'g', metaKey: true }));
    previous?.cleanup?.();
    expect(oldDeactivate).toHaveBeenCalledOnce();

    const activate = vi.fn();
    const current = plugin.setup?.({
      activate,
      deactivate: vi.fn(),
      isActive: () => false,
      isEnabled: () => true,
    });
    shortcutCleanup = current?.cleanup;
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'g', metaKey: true }));

    expect(oldActivate).toHaveBeenCalledOnce();
    expect(activate).toHaveBeenCalledOnce();
    current?.hooks?.onDeactivate?.();
    expect(current?.hooks?.onElementSelect?.(renderRows())).toBeUndefined();
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
