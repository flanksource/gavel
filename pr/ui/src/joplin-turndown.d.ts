declare module '@joplin/turndown' {
  import TurndownService from 'turndown';

  export default TurndownService;
}

declare module '@joplin/turndown-plugin-gfm' {
  import type TurndownService from 'turndown';

  export type TurndownPlugin = (service: TurndownService) => void;

  export const gfm: TurndownPlugin;
  export const highlightedCodeBlock: TurndownPlugin;
  export const strikethrough: TurndownPlugin;
  export const tables: TurndownPlugin;
  export const taskListItems: TurndownPlugin;
}
