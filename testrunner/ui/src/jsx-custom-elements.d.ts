// Declares runtime custom elements used in JSX so tsc --noEmit doesn't fail.
// These elements are registered at runtime by their own scripts (e.g.
// @iconify/iconify-icon) and have no compile-time React component type.

import type { HTMLAttributes } from 'react';

declare module 'react' {
  namespace JSX {
    interface IntrinsicElements {
      'iconify-icon': HTMLAttributes<HTMLElement> & {
        icon?: string;
        width?: string | number;
        height?: string | number;
        flip?: string;
        rotate?: string | number;
        inline?: boolean;
      };
    }
  }
}

// Ensures this file is treated as a module; otherwise the `declare module`
// augmentation above is ignored.
export {};
