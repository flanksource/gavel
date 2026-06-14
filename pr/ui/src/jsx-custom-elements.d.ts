// Declares runtime custom elements used in JSX so tsc --noEmit doesn't fail.
// <iconify-icon> is registered at runtime by the iconify-icon script and has no
// compile-time React component type. React 19 maps `className` to the `class`
// attribute on custom elements, so components use `className` here as usual.

import 'react';

declare module 'react' {
  namespace JSX {
    interface IntrinsicElements {
      'iconify-icon': React.HTMLAttributes<HTMLElement> & {
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
