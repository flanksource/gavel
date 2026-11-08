import { DropdownMenu, type DropdownMenuItem } from "@flanksource/clicky-ui/components";
import { formatRelativeTime } from "@flanksource/clicky-ui/data";
import { UiChatDots } from "@flanksource/clicky-ui/icons";
import type { RecentComment } from "./comment-summary";

/**
 * The navbar's comments indicator: an icon with the unread count that opens the
 * most recent threads across every edited file, newest first. Picking one opens its file.
 */
export function RecentComments({
  recent,
  unread,
  pathOf,
  onOpen,
}: {
  recent: RecentComment[];
  unread: number;
  /** The file's path as the tree shows it. */
  pathOf: (fileId: string) => string;
  onOpen: (comment: RecentComment) => void;
}) {
  const label = unread ? `Comments (${unread} unread)` : "Comments";
  const items: DropdownMenuItem[] = recent.map((comment) => ({
    label: (
      <span className="flex w-full min-w-0 flex-col gap-0.5 py-0.5">
        <span className="flex min-w-0 items-center gap-1.5">
          <span
            className={`inline-block size-1.5 shrink-0 rounded-full ${comment.unread ? "bg-primary" : "bg-transparent"}`}
            aria-label={comment.unread ? "unread" : undefined}
          />
          <span className={`truncate font-mono text-[11px] ${comment.unread ? "font-semibold text-foreground" : "text-muted-foreground"}`}>
            {pathOf(comment.fileId)}
          </span>
          <span className="ml-auto shrink-0 text-[10px] text-muted-foreground">{formatRelativeTime(new Date(comment.at))}</span>
        </span>
        <span className="truncate pl-3 text-xs">
          <span className="font-medium">{comment.author.label ?? comment.author.id}:</span> {comment.body}
        </span>
      </span>
    ),
    title: comment.body,
    onSelect: () => onOpen(comment),
  }));
  if (items.length === 0) items.push({ label: "No comments yet", disabled: true, onSelect: () => {} });
  return (
    <DropdownMenu
      align="right"
      menuLabel="Recent comments"
      menuClassName="w-96"
      header={<span className="text-xs font-semibold">Recent comments</span>}
      items={items}
      trigger={
        <button
          type="button"
          title={label}
          aria-label={label}
          className="relative inline-flex size-8 items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-accent-foreground"
        >
          <UiChatDots className="text-base" />
          {unread > 0 && (
            <span className="pointer-events-none absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-primary px-1 text-[10px] font-semibold leading-none text-primary-foreground">
              {unread}
            </span>
          )}
        </button>
      }
    />
  );
}
