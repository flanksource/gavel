import preset from '@flanksource/clicky-ui/tailwind-preset';

/** @type {import('tailwindcss').Config} */
export default {
  presets: [preset],
  // The testrunner library (@flanksource/gavel/testrunner) ships compiled, but its
  // components' classes are written in its source, which is what gets scanned.
  content: ['./src/**/*.{ts,tsx}', '../../testrunner/ui/src/**/*.{ts,tsx}'],
};
