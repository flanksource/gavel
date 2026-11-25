import { skipToken, useQuery, useQueryClient } from '@tanstack/react-query';
import { Modal, useModal, useToast, type HostedModalProps, type ModalFormDefinition } from '@flanksource/clicky-ui/components';
import { queryKeys } from '../../query';
import type { PRDetail, PRItem, Project } from '../../types';
import { CreateTodoFromPRDialog } from './CreateTodoFromPRDialog';
import { RunCreationRecovery, type RunAfterCreationState } from './RunAfterCreation';
import { readPRTodoDraft, newPRTodoFormKey, prTodoHref, prTodoRunRecoveryFormKey, type PRTodoDraft } from './prTodoDraft';
import type { CreateTodoResponse } from './todoMutations';

function HostedPRTodoDialog({ initialDraft, modalProps, submit }: HostedModalProps) {
  const pr = initialDraft.pr as PRItem;
  const recovered = initialDraft.draft as PRTodoDraft | undefined;
  const initialWorkspaces = initialDraft.workspaces as Project[] | undefined;
  const queryClient = useQueryClient();
  const { openModal } = useModal();
  const { toast } = useToast();
  const { data: detail } = useQuery<PRDetail>({ queryKey: queryKeys.prDetail(pr.repo, pr.number), queryFn: skipToken });
  const { data: projects } = useQuery<Project[]>({ queryKey: queryKeys.projects(), queryFn: skipToken });
  const workspaces = initialWorkspaces ?? projects ?? (recovered ? [{ name: recovered.dir, dir: recovered.dir, repos: [pr.repo] }] : []);
  const refreshProjects = () => { void queryClient.invalidateQueries({ queryKey: queryKeys.projects(), exact: true }); };

  function onResult(result: CreateTodoResponse, run: RunAfterCreationState['submitted']) {
    refreshProjects();
    void queryClient.invalidateQueries({ queryKey: queryKeys.prLinkedTodos(pr.url), exact: true });
    if (result.run?.status === 'failed' && run) {
      openModal(prTodoRunRecoveryFormKey, { result, run });
      return;
    }
    toast({
      tone: 'success',
      message: <>Todo created. <a className="underline" href={prTodoHref(result.todo.ref)}>Open todo</a></>,
    });
  }

  return <CreateTodoFromPRDialog
    hosted={{ initialDraft, modalProps, submit }}
    recovered={recovered}
    pr={pr}
    detail={detail ?? (initialDraft.detail as PRDetail | undefined) ?? null}
    workspaces={workspaces}
    onProjectsChanged={refreshProjects}
    onResult={onResult}
  />;
}

function HostedPRTodoRunRecovery({ initialDraft, modalProps }: HostedModalProps) {
  return <Modal {...modalProps}>
    <RunCreationRecovery result={initialDraft.result as CreateTodoResponse}
      request={initialDraft.run as RunAfterCreationState['submitted']} onDone={modalProps.onClose} />
  </Modal>;
}

export const prTodoForms = {
  [newPRTodoFormKey]: {
    title: 'New todo from PR',
    Form: HostedPRTodoDialog,
    recovery: {
      serialize: draft => {
        readPRTodoDraft(draft);
        if (typeof draft.snapshot !== 'string') throw new Error('Missing PR todo snapshot');
        return { snapshot: draft.snapshot };
      },
      restore: readPRTodoDraft,
    },
  },
  [prTodoRunRecoveryFormKey]: { title: 'Todo created', Form: HostedPRTodoRunRecovery },
} satisfies Record<string, ModalFormDefinition>;
