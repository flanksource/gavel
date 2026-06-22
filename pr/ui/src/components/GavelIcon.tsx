import type { SVGProps } from 'react';
import {
  UiActivity,
  UiAdd,
  UiArrowDown,
  UiArrowLeft,
  UiBeaker,
  UiCancel,
  UiCheck,
  UiCheckFilled,
  UiChevronDown,
  UiChevronRight,
  UiChevronUp,
  UiCircleFilled,
  UiCircleXFilled,
  UiClock,
  UiClose,
  UiCloudDownload,
  UiCog,
  UiComment,
  UiCopy,
  UiDatabase,
  UiDebugStepOver,
  UiDesktop,
  UiDiff,
  UiError,
  UiEye,
  UiEyeClosed,
  UiFolder,
  UiFullscreen,
  UiGitBranch,
  UiGitGraph,
  UiGitMerge,
  UiGitPr,
  UiGlobe,
  UiGraph,
  UiHubot,
  UiJson,
  UiLayers,
  UiLightbulb,
  UiLinkExternal,
  UiLoader,
  UiMarkdown,
  UiMoon,
  UiOrganization,
  UiPass,
  UiPause,
  UiPlay,
  UiRefresh,
  UiRestart,
  UiRobotAi,
  UiSearch,
  UiServerProcess,
  UiStop,
  UiSun,
  UiSync,
  UiTrash,
  UiUnknown,
  UiWarningTriangle,
  UiWatch,
} from '@flanksource/clicky-ui/icons';

type StaticIcon = any;

function VercelIcon({ size = '1em', className, title, ...props }: SVGProps<SVGSVGElement> & { size?: number | string; title?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={size}
      height={size}
      viewBox="0 0 24 24"
      role={title ? 'img' : 'presentation'}
      aria-label={title}
      aria-hidden={title ? undefined : true}
      className={className}
      {...props}
    >
      <path fill="currentColor" d="M12 3 23 21H1L12 3Z" />
    </svg>
  );
}

const icons: Record<string, StaticIcon> = {
  'codicon:add': UiAdd,
  'codicon:arrow-down': UiArrowDown,
  'codicon:arrow-left': UiArrowLeft,
  'codicon:beaker': UiBeaker,
  'codicon:beaker-stop': UiBeaker,
  'codicon:check': UiCheck,
  'codicon:chevron-down': UiChevronDown,
  'codicon:chevron-right': UiChevronRight,
  'codicon:chevron-up': UiChevronUp,
  'codicon:circle-slash': UiCancel,
  'codicon:clock': UiClock,
  'codicon:close': UiClose,
  'codicon:cloud-download': UiCloudDownload,
  'codicon:color-mode': UiDesktop,
  'codicon:comment-discussion': UiComment,
  'codicon:copy': UiCopy,
  'codicon:database': UiDatabase,
  'codicon:debug-pause': UiPause,
  'codicon:debug-restart': UiRestart,
  'codicon:debug-start': UiPlay,
  'codicon:debug-step-over': UiDebugStepOver,
  'codicon:debug-stop': UiStop,
  'codicon:device-desktop': UiDesktop,
  'codicon:diff': UiDiff,
  'codicon:error': UiError,
  'codicon:eye': UiEye,
  'codicon:eye-closed': UiEyeClosed,
  'codicon:folder': UiFolder,
  'codicon:gear': UiCog,
  'codicon:git-branch': UiGitBranch,
  'codicon:git-merge': UiGitMerge,
  'codicon:git-pull-request': UiGitPr,
  'codicon:globe': UiGlobe,
  'codicon:graph': UiGraph,
  'codicon:hubot': UiHubot,
  'codicon:json': UiJson,
  'codicon:layers': UiLayers,
  'codicon:lightbulb': UiLightbulb,
  'codicon:link-external': UiLinkExternal,
  'codicon:markdown': UiMarkdown,
  'codicon:moon': UiMoon,
  'codicon:organization': UiOrganization,
  'codicon:pass': UiPass,
  'codicon:play': UiPlay,
  'codicon:pulse': UiActivity,
  'codicon:refresh': UiRefresh,
  'codicon:screen-full': UiFullscreen,
  'codicon:search': UiSearch,
  'codicon:server-process': UiServerProcess,
  'codicon:sun': UiSun,
  'codicon:sync': UiSync,
  'codicon:sync-ignored': UiSync,
  'codicon:trash': UiTrash,
  'codicon:warning': UiWarningTriangle,
  'codicon:watch': UiWatch,
  'octicon:alert-fill-16': UiWarningTriangle,
  'octicon:check-circle-fill-16': UiCheckFilled,
  'octicon:dot-fill-16': UiCircleFilled,
  'octicon:git-merge-16': UiGitMerge,
  'octicon:git-pull-request-16': UiGitPr,
  'octicon:git-pull-request-closed-16': UiGitPr,
  'octicon:git-pull-request-draft-16': UiGitPr,
  'octicon:graph-16': UiGitGraph,
  'octicon:x-circle-fill-16': UiCircleXFilled,
  'simple-icons:githubcopilot': UiRobotAi,
  'simple-icons:vercel': VercelIcon,
  'svg-spinners:ring-resize': UiLoader,
};

const spinning = new Set(['svg-spinners:ring-resize']);

function cx(...classes: Array<string | false | undefined>) {
  return classes.filter(Boolean).join(' ');
}

export interface GavelIconProps extends Omit<SVGProps<SVGSVGElement>, 'name'> {
  name: string;
  size?: number | string;
  title?: string;
}

export function GavelIcon({ name, className, ...props }: GavelIconProps) {
  const Icon = icons[name] ?? UiUnknown;
  return (
    <Icon
      {...props}
      className={cx('inline-block shrink-0 align-[-0.125em]', spinning.has(name) && 'animate-spin', className)}
    />
  );
}
