import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import {
  Button,
  CommentProvider,
  CommentSidePanel,
  DensityProvider,
  DOCUMENT_ANCHOR,
  Tabs,
  ThemeProvider,
  isUnresolved,
  useCommentContext,
  useCommentRailWidth,
  type Comment,
  type CommentConfig,
  type CommentCreateInput,
  type CommentScrollOptions,
  type DropdownMenuItem,
} from "@flanksource/clicky-ui";
import { CommentCapture, type CommentAgentRun } from "@flanksource/clicky-ui/comments";
import { AgentIdentityProvider } from "@flanksource/clicky-ui/chat";
import { UiLayoutWindow, UiMinus, UiSidebar } from "@flanksource/clicky-ui/icons";
import type { AgentPin } from "@flanksource/page-runtime/schema";
import { AgentPicker, type AgentPickerProps } from "../agents/AgentPicker";
import { commentPins } from "../comments/comment-pins";
import { AgentRunActions } from "./AgentRunActions";
import { freshTarget, retryTarget, withEdits } from "./agent-retry";
import { FloatingDraft } from "../host/FloatingDraft";
import { externalUrl, type AnchorState, type HostState, type Presentation, type Selection } from "../host/model";
import { requestPanelMode, saveHostSession, sendToHost, useHostBridge } from "../host/useHostBridge";
import { useHostDialog } from "../host/useHostDialog";
import { RunAgentDialog } from "../agents/RunAgentDialog";
import { agentAssignItems, assignableRoots } from "./assign-items";
import { deliveryLabel, DraftDeliveryPreview, useDeliveryMentionables, useDeliveryPreview } from "./DeliveryPreview";
import type { CreateDeliveryChoice } from "../comments/gavel-delivery";
import { GavelClient, fetchRuntimeCatalog, type AgentSession, type DeliverTarget } from "./gavel-client";
import { EventSourceProvider } from "@flanksource/clicky-ui/hooks";
import { gavelEventHub } from "./gavel-streams";
import { pageSnapshot } from "./page-snapshot";
import { PanelComposer } from "./PanelComposer";
import { PanelSettingsMenu, type AssignAllAction } from "./PanelSettingsMenu";
import { pickerSessions } from "./picker-sessions";
import { SessionTab, SessionTabLabel } from "./SessionTab";
import type { PanelTarget } from "./settings";
import { useAdoptRunSession, useAgentSessions, useAutoRun, useRuntimeCatalog } from "./useGavelAgents";
import { GAVEL_COMMENT_CONFIG, useGavelComments, type GavelComments } from "./useGavelComments";
import styles from "./GavelPanel.module.css";

let detailOpen = false;

const TAB_KEY = "gavel.panel.tab";
const PAGE_SNAPSHOT_TIMEOUT_MS = 10_000;
type PanelTab = "comments" | "session";

function maximizeHost(open: boolean) {
  detailOpen = open;
  const stored = localStorage.getItem("playground.bookmarklet.width");
  sendToHost({ type: "sidebar-layout", width: stored === null ? 400 : Number(stored), maximized: open });
}

const panelHost = { send: sendToHost, requestMode: requestPanelMode, maximize: maximizeHost };

function describe(cause: unknown): string {
  return cause instanceof Error ? cause.message : String(cause);
}

export function GavelPanel({ target }: { target: PanelTarget }) {
  const client = useMemo(() => new GavelClient({ server: target.server, project: target.project }), [target.server, target.project]);
  // Every SSE stream the panel's components open (the session transcript's live follow) shares one hub connection to gavel.
  const streams = useMemo(() => gavelEventHub(target.server), [target.server]);
  const bridge = useHostBridge();
  const [sidebarWidth] = useCommentRailWidth(400, "playground.bookmarklet.width");
  useEffect(() => {
    sendToHost({ type: "sidebar-layout", width: sidebarWidth, maximized: detailOpen });
  }, [sidebarWidth]);
  const page = bridge.state ? { url: bridge.state.url, title: bridge.state.title } : undefined;
  const comments = useGavelComments({ client, page });
  const [agentSession, setAgentSession] = useState(target.agentSession);
  const [deliveryError, setDeliveryError] = useState("");
  const { sessions, error: sessionsError, refresh: refreshSessions } = useAgentSessions(client);
  const identities = useMemo(() => Object.fromEntries((sessions ?? []).map(session => [session.id, { sessionName: session.title, ...(session.model ? { model: session.model } : {}) }])), [sessions]);
  // The runtime for new sessions starts as the one the grabber was opened with; changing it here applies to this panel until it is closed.
  const [runtime, setRuntime] = useState(target.runtime);
  const { catalog, error: catalogError } = useRuntimeCatalog(target.server);
  const runDelivery = page ? client.runTarget({ page, grouping: target.grouping, ...(runtime ? { runtime } : {}) }) : undefined;
  const [handingOff, setHandingOff] = useState(false);
  const autoRun = useAutoRun({ client, page, run: runDelivery });
  const defaultState = useDeliveryPreview({ client, page, sessionId: agentSession, revision: autoRun.agent, signal: sessions });
  const delivery = defaultState.preview?.route?.target;
  const resolvedSession = defaultState.preview?.session;
  const pageUrl = page?.url;
  const pageRef = useRef(pageUrl);
  pageRef.current = pageUrl;
  const selectionRevision = useRef(0);
  const pendingAdoption = useRef<{ page: string; revision: number } | undefined>(undefined);
  const sessionWrites = useRef(Promise.resolve());
  const [savingSession, setSavingSession] = useState(false);
  const targets = useDeliveryMentionables(client, page, sessions);

  const chooseSession = (session: string | undefined) => {
    const revision = ++selectionRevision.current;
    setSavingSession(true);
    setDeliveryError("");
    sessionWrites.current = sessionWrites.current.then(async () => {
      try {
        await saveHostSession(session);
        setAgentSession(session);
      } catch (cause) {
        if (revision === selectionRevision.current) setDeliveryError(describe(cause));
      } finally {
        if (revision === selectionRevision.current) setSavingSession(false);
      }
    });
  };

  const adoption = useAdoptRunSession({
    comments: comments.comments,
    adopt: sessionId => {
      const requested = pendingAdoption.current;
      pendingAdoption.current = undefined;
      if (!requested || pageRef.current !== requested.page || selectionRevision.current !== requested.revision) return;
      chooseSession(sessionId);
      void refreshSessions();
    },
  });

  async function handOff(choose: () => DeliverTarget, send: (target: DeliverTarget) => Promise<unknown>) {
    setDeliveryError("");
    setHandingOff(true);
    try {
      await send(choose());
    } catch (cause) {
      setDeliveryError(describe(cause));
    } finally {
      setHandingOff(false);
      void refreshSessions();
    }
  }
  const deliverTo = (comment: Comment, choose: () => DeliverTarget) => handOff(
    () => withEdits(choose(), autoRun.agent?.agent.edits), target => comments.deliver(comment.id, target),
  );
  const pageNotReported = () => new Error("The grabber has not reported the page yet");
  const sendFresh = (comment: Comment) => deliverTo(comment, () => {
    if (!runDelivery) throw pageNotReported();
    return freshTarget(runDelivery);
  });
  const retry = (comment: Comment, run: CommentAgentRun) => deliverTo(comment, () => {
    if (!delivery) throw new Error("Choose an available delivery target in Settings");
    return retryTarget({ run, delivery });
  });
  const config: CommentConfig = {
    ...GAVEL_COMMENT_CONFIG,
    targets: targets?.items ?? [],
    renderDeliveryPreview: draft => <DraftDeliveryPreview client={client} page={page} sessionId={agentSession}
      revision={autoRun.agent} defaultState={defaultState} body={draft.body}
      {...(draft.target ? { mentionId: draft.target.id } : {})} />,
  };
  const create = async (input: CommentCreateInput, selection: Selection) => {
    if (!page) throw pageNotReported();
    if (savingSession) throw new Error("Wait for the delivery target to finish saving.");
    const revision = selectionRevision.current;
    const choice: CreateDeliveryChoice = {
      ...(agentSession ? { defaultSessionId: agentSession } : {}),
      ...(input.target ? { mentionId: input.target.id } : {}),
    };
    const preview = await client.previewDelivery({ page, body: input.body, delivery: choice });
    if (preview.route?.target.kind === "session" && !preview.session)
      throw new Error("The delivery session is unavailable. Choose a different target.");
    if (pageRef.current !== page.url || selectionRevision.current !== revision)
      throw new Error("The page or delivery setting changed. Review the target and post again.");
    const id = await comments.create(input, selection, choice);
    if (preview.route?.target.kind === "run" && !preview.session && !preview.mention &&
      pageRef.current === page.url && revision === selectionRevision.current) {
      pendingAdoption.current = { page: page.url, revision };
      adoption.watchNew(id);
    }
    if (preview.route) void refreshSessions();
  };
  const assignable = assignableRoots(comments.comments, resolvedSession?.id);
  const assignAll = {
    label: `Assign All to ${deliveryLabel(defaultState)} (${assignable.length})`,
    disabled: savingSession || handingOff || !delivery || assignable.length === 0 || (delivery.kind === "session" && !resolvedSession),
    onSelect: () => {
      if (assignAll.disabled || !delivery || !page) return;
      const ids = assignable.map(thread => thread.id);
      const revision = selectionRevision.current;
      const launches = delivery.kind === "run" && !resolvedSession;
      if (launches) {
        pendingAdoption.current = { page: page.url, revision };
        adoption.watch(ids);
      }
      void handOff(() => delivery, async target => {
        try {
          await comments.deliverAll(ids, resolvedSession ? { kind: "session", sessionId: resolvedSession.id, ...(target.edits ? { edits: target.edits } : {}) } : target);
        } catch (cause) {
          if (launches) adoption.cancel();
          throw cause;
        }
      });
    },
  };
  const renderRunDetails = (comment: Comment) => {
    const run = comment.agentRun;
    const capture = comments.comments.find(entry => entry.id === comment.id)?.element;
    return (
      <>
        {capture && <CommentCapture capture={capture} />}
        {run?.state === "error" && (
          <>
            <p role="alert" className="px-1 text-xs text-destructive">{run.error ?? "Agent hand-off failed without an error message."}</p>
            <AgentRunActions
              state={run.state}
              {...(comment.status !== undefined ? { status: comment.status } : {})}
              busy={handingOff}
              onRetry={() => void retry(comment, run)}
              onRetryFresh={() => void sendFresh(comment)}
            />
          </>
        )}
      </>
    );
  };

  const picker: AgentPickerProps = {
    heading: "Agent",
    runtime: { value: runtime, catalog, onChange: setRuntime, ariaLabel: "Runtime for new sessions", emptyLabel: "Gavel's default" },
    autoRun: {
      checked: autoRun.enabled,
      disabled: !autoRun.ready,
      label: "Auto-run new comments on this page",
      title: "Use Gavel’s resolved rule when there is no remembered session or draft target.",
      onChange: enabled => void autoRun.set(enabled),
    },
    edits: {
      value: autoRun.edits,
      disabled: !autoRun.ready,
      title: "Whether deliveries edit this project’s files or propose patches for review.",
      onChange: next => void autoRun.setEdits(next),
    },
    sessions: {
      items: pickerSessions({ sessions, selectedId: agentSession }),
      state: sessions ? "ready" : sessionsError ? "error" : "loading",
      selectedId: agentSession,
      onSelect: session => chooseSession(session.id),
      onNew: () => chooseSession(undefined),
      newLabel: "Use automatic target",
      selectLabel: session => `Use session ${session.title}`,
    },
  };
  // The thread whose Advanced… assignment is being chosen.
  const [advanced, setAdvanced] = useState<Comment>();
  const loadCatalog = useCallback(() => fetchRuntimeCatalog({ server: target.server }), [target.server]);
  const error = comments.error ?? bridge.error;
  const alerts = [error, deliveryError, sessionsError, catalogError, autoRun.error, targets?.error].filter((message): message is string => !!message);

  return (
    <ThemeProvider>
      <EventSourceProvider value={streams}>
      <AgentIdentityProvider identities={identities}>
      <DensityProvider>
        <main className="flex h-dvh min-h-0 flex-col bg-background text-foreground">
          {!bridge.state && (
            <>
              <Alerts messages={alerts} />
              {!bridge.error && <p role="status" className="p-3 text-sm">Connecting to this page…</p>}
            </>
          )}
          {bridge.state && (
            <GavelRail
              host={panelHost}
              state={bridge.state}
              comments={comments}
              config={config}
              create={create}
              floating={bridge.floating}
              presentation={bridge.presentation}
              anchors={bridge.anchors}
              {...(bridge.focus ? { focus: bridge.focus } : {})}
              alerts={<Alerts messages={alerts} />}
              renderRunDetails={renderRunDetails}
              session={{ sessionId: resolvedSession?.id, session: resolvedSession, label: deliveryLabel(defaultState) }}
              sessionContent={<SessionTab server={target.server} sessionId={resolvedSession?.id} targetState={defaultState} />}
              settingsContent={<AgentPicker {...picker} />}
              assignAll={assignAll}
              extraAssignItems={comment => comment.parentId ? [] : agentAssignItems({
                sessions: sessions ?? [],
                currentSessionId: comment.agentRun?.sessionId,
                runtime,
                disabled: !isUnresolved(GAVEL_COMMENT_CONFIG, comment.status),
                onSession: session => void deliverTo(comment, () => ({ kind: "session", sessionId: session.id })),
                onNew: () => void sendFresh(comment),
                onAdvanced: () => setAdvanced(comment),
              })}
            />
          )}
        </main>
        <RunAgentDialog
          open={advanced !== undefined}
          mode="assign"
          grouping={target.grouping}
          current={runtime}
          error={deliveryError || null}
          busy={handingOff}
          loadCatalog={loadCatalog}
          onClose={() => setAdvanced(undefined)}
          onRun={chosen => {
            if (!advanced) return;
            void deliverTo(advanced, () => {
              if (!page) throw pageNotReported();
              return freshTarget(client.runTarget({ page, grouping: target.grouping, runtime: chosen }));
            }).then(() => setAdvanced(undefined));
          }}
        />
      </DensityProvider>
      </AgentIdentityProvider>
      </EventSourceProvider>
    </ThemeProvider>
  );
}

function Alerts({ messages }: { messages: string[] }) {
  if (messages.length === 0) return null;
  return (
    <div role="alert" className="shrink-0 space-y-1 border-b border-border p-3 text-sm text-destructive">
      {messages.map(message => <p key={message}>{message}</p>)}
    </div>
  );
}

export type GavelRailProps = {
  host: typeof panelHost;
  state: HostState;
  comments: GavelComments;
  /** The comment config with this panel's targets. */
  config: CommentConfig;
  /** Creates a thread and delivers it to the target its composer chose. */
  create: (input: CommentCreateInput, selection: Selection) => Promise<void>;
  floating: boolean;
  presentation: Presentation;
  anchors: AnchorState[];
  focus?: string;
  alerts: ReactNode;
  renderRunDetails: (comment: Comment) => ReactNode;
  session: { sessionId: string | undefined; session: AgentSession | undefined; label?: string };
  sessionContent: ReactNode;
  settingsContent: ReactNode;
  extraAssignItems: (comment: Comment) => DropdownMenuItem[];
  /** The page-level "assign every open thread" entries. */
  assignAll: AssignAllAction;
};

export function GavelRail(props: GavelRailProps) {
  const { state, comments, anchors } = props;
  const navigate = (anchor: string, options: CommentScrollOptions) => {
    if (!anchors.some(value => value.anchor === anchor && value.available)) return false;
    props.host.send({ type: "scroll", anchor, ...options });
    return true;
  };
  return (
    <CommentProvider
      persistenceKey={`gavel:${externalUrl(state.url).base}`}
      comments={comments.comments}
      config={props.config}
      initialRailMode="all"
      onCreate={async input => {
        const selection = state.selection;
        if (!selection || input.anchor !== selection.anchor || externalUrl(selection.element.pageUrl ?? "").url !== externalUrl(state.url).url)
          throw new Error("Select an element or Comment on page on the current URL before posting this draft");
        await props.create(input, selection);
      }}
      onReply={comments.reply}
      onUpdateStatus={comments.updateStatus}
      onClose={comments.close}
      onAcknowledge={comments.acknowledge}
      onAssign={comments.assign}
      onDelete={comments.remove}
      onApplyProposal={comments.applyProposal}
      onRejectProposal={comments.rejectProposal}
      onRequestRevision={comments.requestRevision}
      loadProposalDiff={comments.loadProposalDiff}
      onScrollToAnchor={navigate}
      getAnchorTop={anchor => anchors.find(value => value.anchor === anchor && value.available)?.top ?? null}
    >
      <RailContents {...props} />
    </CommentProvider>
  );
}

function readTab(): PanelTab {
  const stored = localStorage.getItem(TAB_KEY);
  return stored === "session" ? "session" : "comments";
}

function RailContents({ state, comments, config, create, floating, presentation, anchors, focus, alerts, renderRunDetails, session, sessionContent, settingsContent, extraAssignItems, assignAll, host }: GavelRailProps) {
  const ctx = useCommentContext();
  const [tab, setTab] = useState<PanelTab>(readTab);
  const [snapshot] = useState(() => pageSnapshot({ requestPage: () => host.send({ type: "page" }), timeoutMs: PAGE_SNAPSHOT_TIMEOUT_MS }));
  const chooseTab = (next: string) => {
    if (next !== "comments" && next !== "session") throw new Error(`Unknown sidebar tab: ${next}`);
    setTab(next);
    localStorage.setItem(TAB_KEY, next);
  };
  useHostDialog(true, host.send);
  const labels = Object.fromEntries(anchors.map(value => [value.anchor, value.label]));
  if (state.selection) labels[state.selection.anchor] = state.selection.label;
  const pins = JSON.stringify(commentPins(comments.comments.filter(comment => ctx.comments.some(visible => visible.id === comment.id))));
  useEffect(() => {
    host.send({ type: "pins", anchors: JSON.parse(pins) as AgentPin[] });
  }, [pins, state, host]);
  useEffect(() => {
    if (!state.selection) return;
    // A snapshot the footer composer asked for posts from there instead of opening a floating draft.
    if (snapshot.claim(state.selection)) return;
    ctx.focusAnchor(state.selection.anchor);
    host.send({ type: "draft", active: true, anchor: state.selection.anchor });
  }, [state.selection, ctx.focusAnchor, host]);
  useEffect(() => {
    if (focus) { ctx.focusAnchor(focus); chooseTab("comments"); }
  }, [focus, ctx.focusAnchor]);
  useEffect(() => {
    host.send({ type: "highlight", anchor: ctx.highlightAnchor });
  }, [ctx.highlightAnchor, host]);
  if (floating && state.selection) return <FloatingDraft selection={state.selection} />;

  const openThreads = comments.comments.filter(comment => !comment.parentId && isUnresolved(GAVEL_COMMENT_CONFIG, comment.status)).length;
  const postPageComment = async (input: CommentCreateInput) => {
    const selection = await snapshot.request();
    await create({ ...input, anchor: DOCUMENT_ANCHOR }, selection);
  };
  const rail = presentation === "rail";
  return (
    <>
      <header className="flex shrink-0 items-center gap-1 border-b border-border pr-1" title={state.url}>
        <Tabs
          className={`min-w-0 flex-1 gap-0 border-b-0 ${styles.tabs}`}
          value={tab}
          onChange={chooseTab}
          tabs={[
            { id: "comments", label: "Comments", count: openThreads },
            { id: "session", label: <SessionTabLabel sessionId={session.sessionId} session={session.session} {...(session.label ? { label: session.label } : {})} /> },
          ]}
        />
        <PanelSettingsMenu settingsContent={settingsContent} assignAll={assignAll} />
        <Button
          size="icon"
          variant="ghost"
          className="size-7 shrink-0"
          aria-label={rail ? "Show as hover panel" : "Dock as side rail"}
          title={rail ? "Show as hover panel" : "Dock as side rail"}
          onClick={() => host.requestMode(rail ? "compact" : "rail")}
        >
          {rail ? <UiLayoutWindow className="size-4" aria-hidden /> : <UiSidebar className="size-4" aria-hidden />}
        </Button>
        <Button size="icon" variant="ghost" className="size-7 shrink-0" aria-label="Minimize" title="Minimize" onClick={() => host.send({ type: "minimize" })}>
          <UiMinus className="size-4" aria-hidden />
        </Button>
      </header>
      {alerts}
      {tab === "session" ? (
        // The inspector scrolls its own transcript under a fixed toolbar, so this body does not scroll.
        <div className="flex min-h-0 flex-1 flex-col">
          {sessionContent}
        </div>
      ) : (
        <>
          <div className="min-h-0 flex-1 overflow-y-auto p-2">
            <CommentSidePanel
              previewLines={3}
              onMaximizedChange={host.maximize}
              compact
              hideComposer
              focusedAlignment="flow"
              className="!w-full"
              anchorLabels={labels}
              formatAnchorLabel={anchor => anchor}
              extraAssignItems={extraAssignItems}
              renderRootDetails={renderRunDetails}
              detailTabs={[{
                id: "capture",
                label: "Capture",
                render: thread => <CommentCapture capture={comments.comments.find(comment => comment.id === thread[0]?.id)?.element} expanded />,
              }]}
            />
            {anchors.some(anchor => !anchor.available) && (
              <p role="status" className="mt-3 text-xs text-muted-foreground">Some anchors are unavailable. Their comments are retained.</p>
            )}
          </div>
          <PanelComposer
            config={config}
            onPost={postPageComment}
            onPick={({ screenshot }) => host.send({ type: "pick", ...(screenshot ? { screenshot } : {}) })}
          />
        </>
      )}
    </>
  );
}
