// Issues icon set — Phosphor (`ph:*-thin`) glyphs bundled offline via @iconify/react.
// Each icon carries its semantic default color (overridable with the `color` prop).
// Metadata for every icon lives in ISSUE_ICONS, keyed by its spec key.
import { Icon, type IconProps } from '@iconify/react/offline';

import sirenThin from '@iconify-icons/ph/siren-thin';
import warningOctagonThin from '@iconify-icons/ph/warning-octagon-thin';
import warningThin from '@iconify-icons/ph/warning-thin';
import warningCircleThin from '@iconify-icons/ph/warning-circle-thin';
import infoThin from '@iconify-icons/ph/info-thin';
import circleThin from '@iconify-icons/ph/circle-thin';
import funnelThin from '@iconify-icons/ph/funnel-thin';
import circleHalfThin from '@iconify-icons/ph/circle-half-thin';
import eyeThin from '@iconify-icons/ph/eye-thin';
import prohibitThin from '@iconify-icons/ph/prohibit-thin';
import checkCircleThin from '@iconify-icons/ph/check-circle-thin';
import archiveThin from '@iconify-icons/ph/archive-thin';
import arrowCounterClockwiseThin from '@iconify-icons/ph/arrow-counter-clockwise-thin';
import xCircleThin from '@iconify-icons/ph/x-circle-thin';
import copyThin from '@iconify-icons/ph/copy-thin';
import bugThin from '@iconify-icons/ph/bug-thin';
import fireThin from '@iconify-icons/ph/fire-thin';
import sparkleThin from '@iconify-icons/ph/sparkle-thin';
import trendUpThin from '@iconify-icons/ph/trend-up-thin';
import checkSquareThin from '@iconify-icons/ph/check-square-thin';
import bookmarkSimpleThin from '@iconify-icons/ph/bookmark-simple-thin';
import lightningThin from '@iconify-icons/ph/lightning-thin';
import treeStructureThin from '@iconify-icons/ph/tree-structure-thin';
import questionThin from '@iconify-icons/ph/question-thin';
import arrowFatLinesUpThin from '@iconify-icons/ph/arrow-fat-lines-up-thin';
import arrowFatUpThin from '@iconify-icons/ph/arrow-fat-up-thin';
import equalsThin from '@iconify-icons/ph/equals-thin';
import arrowFatDownThin from '@iconify-icons/ph/arrow-fat-down-thin';
import dotsThreeThin from '@iconify-icons/ph/dots-three-thin';
import userThin from '@iconify-icons/ph/user-thin';
import userCircleThin from '@iconify-icons/ph/user-circle-thin';
import tagThin from '@iconify-icons/ph/tag-thin';
import flagBannerThin from '@iconify-icons/ph/flag-banner-thin';
import arrowsClockwiseThin from '@iconify-icons/ph/arrows-clockwise-thin';
import gaugeThin from '@iconify-icons/ph/gauge-thin';
import calendarThin from '@iconify-icons/ph/calendar-thin';
import chatCircleThin from '@iconify-icons/ph/chat-circle-thin';
import paperclipThin from '@iconify-icons/ph/paperclip-thin';
import bellThin from '@iconify-icons/ph/bell-thin';
import linkThin from '@iconify-icons/ph/link-thin';
import linkBreakThin from '@iconify-icons/ph/link-break-thin';

export type IssueGroup = 'severity' | 'status' | 'type' | 'priority' | 'meta';

export interface IssueIconMeta {
  key: string;
  name: string;
  group: IssueGroup;
  color: string;
  meaning: string;
  description: string;
}

const ACT_NOW = 'Critical / blocking — act immediately';
const THIS_SHIFT = 'High urgency — act this shift';
const NEUTRAL = 'Neutral / inactive metadata';
const INFO = 'Low / informational';
const ACTIONABLE = 'Open / actionable / linked';
const CLASSIFY = 'Triage & new scope (classify)';
const IN_FLIGHT = 'In flight — actively changing';
const REVIEW = 'Under review / large initiative';
const DELIVERED = 'Resolved / value delivered';

export const ISSUE_ICONS: Record<string, IssueIconMeta> = {
  critical: { key: 'critical', name: 'Critical', group: 'severity', color: '#e11d48', meaning: ACT_NOW, description: 'Service down or data loss. Page on-call now.' },
  high: { key: 'high', name: 'High', group: 'severity', color: '#ea580c', meaning: THIS_SHIFT, description: 'Major degradation; needs attention this shift.' },
  medium: { key: 'medium', name: 'Medium', group: 'severity', color: '#ca8a04', meaning: 'Moderate — non-urgent attention', description: 'Partial or non-urgent impact.' },
  low: { key: 'low', name: 'Low', group: 'severity', color: '#0ea5e9', meaning: INFO, description: 'Minor issue, no material impact.' },
  trivial: { key: 'trivial', name: 'Trivial', group: 'severity', color: '#6b7280', meaning: NEUTRAL, description: 'Cosmetic or informational only.' },

  open: { key: 'open', name: 'Open', group: 'status', color: '#3578e5', meaning: ACTIONABLE, description: 'Filed and awaiting triage or pickup.' },
  triage: { key: 'triage', name: 'Triage', group: 'status', color: '#7c3aed', meaning: CLASSIFY, description: 'Being categorised, deduped, and prioritised.' },
  in_progress: { key: 'in_progress', name: 'In progress', group: 'status', color: '#d97706', meaning: IN_FLIGHT, description: 'Actively being worked.' },
  in_review: { key: 'in_review', name: 'In review', group: 'status', color: '#6366f1', meaning: REVIEW, description: 'Fix proposed; under review.' },
  blocked: { key: 'blocked', name: 'Blocked', group: 'status', color: '#e11d48', meaning: ACT_NOW, description: 'Waiting on a dependency or decision.' },
  resolved: { key: 'resolved', name: 'Resolved', group: 'status', color: '#059669', meaning: DELIVERED, description: 'Fixed and verified.' },
  closed: { key: 'closed', name: 'Closed', group: 'status', color: '#6b7280', meaning: NEUTRAL, description: 'Done and archived; no further action.' },
  reopened: { key: 'reopened', name: 'Reopened', group: 'status', color: '#d97706', meaning: IN_FLIGHT, description: 'Recurred or regressed; back in flow.' },
  wont_fix: { key: 'wont_fix', name: "Won't fix", group: 'status', color: '#6b7280', meaning: NEUTRAL, description: 'Acknowledged but will not be addressed.' },
  duplicate: { key: 'duplicate', name: 'Duplicate', group: 'status', color: '#6b7280', meaning: NEUTRAL, description: 'Already tracked by another issue.' },

  bug: { key: 'bug', name: 'Bug', group: 'type', color: '#e11d48', meaning: ACT_NOW, description: 'Defect — something behaves incorrectly.' },
  incident: { key: 'incident', name: 'Incident', group: 'type', color: '#ea580c', meaning: THIS_SHIFT, description: 'Live operational event tied to on-call.' },
  feature: { key: 'feature', name: 'Feature', group: 'type', color: '#7c3aed', meaning: CLASSIFY, description: 'New capability or user-facing addition.' },
  improvement: { key: 'improvement', name: 'Improvement', group: 'type', color: '#0d9488', meaning: 'Improvement / enhancement', description: 'Enhancement to existing behaviour.' },
  task: { key: 'task', name: 'Task', group: 'type', color: '#3578e5', meaning: ACTIONABLE, description: 'Discrete unit of work to complete.' },
  story: { key: 'story', name: 'Story', group: 'type', color: '#059669', meaning: DELIVERED, description: 'User-facing increment of value.' },
  epic: { key: 'epic', name: 'Epic', group: 'type', color: '#6366f1', meaning: REVIEW, description: 'Large body of work spanning many issues.' },
  sub_task: { key: 'sub_task', name: 'Sub-task', group: 'type', color: '#6b7280', meaning: NEUTRAL, description: 'Child of a parent issue.' },
  question: { key: 'question', name: 'Question', group: 'type', color: '#0ea5e9', meaning: INFO, description: 'Support or clarification request.' },

  urgent: { key: 'urgent', name: 'Urgent', group: 'priority', color: '#e11d48', meaning: ACT_NOW, description: 'Drop everything; expedite.' },
  p_high: { key: 'p_high', name: 'High', group: 'priority', color: '#ea580c', meaning: THIS_SHIFT, description: 'Schedule ahead of normal work.' },
  normal: { key: 'normal', name: 'Normal', group: 'priority', color: '#0ea5e9', meaning: INFO, description: 'Default queue order.' },
  p_low: { key: 'p_low', name: 'Low', group: 'priority', color: '#6b7280', meaning: NEUTRAL, description: 'Pick up when capacity allows.' },
  no_priority: { key: 'no_priority', name: 'No priority', group: 'priority', color: '#6b7280', meaning: NEUTRAL, description: 'Not yet prioritised.' },

  assignee: { key: 'assignee', name: 'Assignee', group: 'meta', color: '#3578e5', meaning: ACTIONABLE, description: 'Person responsible for the issue.' },
  reporter: { key: 'reporter', name: 'Reporter', group: 'meta', color: '#6b7280', meaning: NEUTRAL, description: 'Person who filed the issue.' },
  label: { key: 'label', name: 'Label', group: 'meta', color: '#6b7280', meaning: NEUTRAL, description: 'Free-form tag for filtering.' },
  milestone: { key: 'milestone', name: 'Milestone', group: 'meta', color: '#3578e5', meaning: ACTIONABLE, description: 'Target release or goal.' },
  sprint: { key: 'sprint', name: 'Sprint', group: 'meta', color: '#3578e5', meaning: ACTIONABLE, description: 'Iteration the issue is committed to.' },
  estimate: { key: 'estimate', name: 'Estimate', group: 'meta', color: '#6b7280', meaning: NEUTRAL, description: 'Story points or effort estimate.' },
  due_date: { key: 'due_date', name: 'Due date', group: 'meta', color: '#6b7280', meaning: NEUTRAL, description: 'When the issue is due.' },
  comment: { key: 'comment', name: 'Comment', group: 'meta', color: '#6b7280', meaning: NEUTRAL, description: 'Discussion thread on the issue.' },
  attachment: { key: 'attachment', name: 'Attachment', group: 'meta', color: '#6b7280', meaning: NEUTRAL, description: 'File or log attached to the issue.' },
  watch: { key: 'watch', name: 'Watch', group: 'meta', color: '#6b7280', meaning: NEUTRAL, description: 'Subscribe to issue updates.' },
  linked: { key: 'linked', name: 'Linked', group: 'meta', color: '#3578e5', meaning: ACTIONABLE, description: 'Relates-to another issue.' },
  blocks: { key: 'blocks', name: 'Blocks', group: 'meta', color: '#e11d48', meaning: ACT_NOW, description: 'Blocks or is blocked by another issue.' },
};

export type IssueIconProps = Omit<IconProps, 'icon'>;

function make(data: IconProps['icon'], color: string) {
  const C = ({ color: override, ...rest }: IssueIconProps) => (
    <Icon icon={data} color={override ?? color} {...rest} />
  );
  return C;
}

// Severity
export const SeverityCritical = make(sirenThin, ISSUE_ICONS.critical.color);
export const SeverityHigh = make(warningOctagonThin, ISSUE_ICONS.high.color);
export const SeverityMedium = make(warningThin, ISSUE_ICONS.medium.color);
export const SeverityLow = make(warningCircleThin, ISSUE_ICONS.low.color);
export const SeverityTrivial = make(infoThin, ISSUE_ICONS.trivial.color);

// Status
export const StatusOpen = make(circleThin, ISSUE_ICONS.open.color);
export const StatusTriage = make(funnelThin, ISSUE_ICONS.triage.color);
export const StatusInProgress = make(circleHalfThin, ISSUE_ICONS.in_progress.color);
export const StatusInReview = make(eyeThin, ISSUE_ICONS.in_review.color);
export const StatusBlocked = make(prohibitThin, ISSUE_ICONS.blocked.color);
export const StatusResolved = make(checkCircleThin, ISSUE_ICONS.resolved.color);
export const StatusClosed = make(archiveThin, ISSUE_ICONS.closed.color);
export const StatusReopened = make(arrowCounterClockwiseThin, ISSUE_ICONS.reopened.color);
export const StatusWontFix = make(xCircleThin, ISSUE_ICONS.wont_fix.color);
export const StatusDuplicate = make(copyThin, ISSUE_ICONS.duplicate.color);

// Issue type
export const TypeBug = make(bugThin, ISSUE_ICONS.bug.color);
export const TypeIncident = make(fireThin, ISSUE_ICONS.incident.color);
export const TypeFeature = make(sparkleThin, ISSUE_ICONS.feature.color);
export const TypeImprovement = make(trendUpThin, ISSUE_ICONS.improvement.color);
export const TypeTask = make(checkSquareThin, ISSUE_ICONS.task.color);
export const TypeStory = make(bookmarkSimpleThin, ISSUE_ICONS.story.color);
export const TypeEpic = make(lightningThin, ISSUE_ICONS.epic.color);
export const TypeSubTask = make(treeStructureThin, ISSUE_ICONS.sub_task.color);
export const TypeQuestion = make(questionThin, ISSUE_ICONS.question.color);

// Priority
export const PriorityUrgent = make(arrowFatLinesUpThin, ISSUE_ICONS.urgent.color);
export const PriorityHigh = make(arrowFatUpThin, ISSUE_ICONS.p_high.color);
export const PriorityNormal = make(equalsThin, ISSUE_ICONS.normal.color);
export const PriorityLow = make(arrowFatDownThin, ISSUE_ICONS.p_low.color);
export const PriorityNone = make(dotsThreeThin, ISSUE_ICONS.no_priority.color);

// Relations & metadata
export const MetaAssignee = make(userThin, ISSUE_ICONS.assignee.color);
export const MetaReporter = make(userCircleThin, ISSUE_ICONS.reporter.color);
export const MetaLabel = make(tagThin, ISSUE_ICONS.label.color);
export const MetaMilestone = make(flagBannerThin, ISSUE_ICONS.milestone.color);
export const MetaSprint = make(arrowsClockwiseThin, ISSUE_ICONS.sprint.color);
export const MetaEstimate = make(gaugeThin, ISSUE_ICONS.estimate.color);
export const MetaDueDate = make(calendarThin, ISSUE_ICONS.due_date.color);
export const MetaComment = make(chatCircleThin, ISSUE_ICONS.comment.color);
export const MetaAttachment = make(paperclipThin, ISSUE_ICONS.attachment.color);
export const MetaWatch = make(bellThin, ISSUE_ICONS.watch.color);
export const MetaLinked = make(linkThin, ISSUE_ICONS.linked.color);
export const MetaBlocks = make(linkBreakThin, ISSUE_ICONS.blocks.color);
