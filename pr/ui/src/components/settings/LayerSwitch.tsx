import { SegmentedControl } from '@flanksource/clicky-ui/components';
import type { SegmentedOption } from '@flanksource/clicky-ui/components';
import { UiCog, UiFolderGit } from '@flanksource/clicky-ui/icons';
import type { SettingsLayer } from './provenance';

// clicky-ui ships React 18 types; our icon components are typed against React 19,
// so the structurally-identical icon prop needs a cast at this boundary.
type LayerIcon = SegmentedOption<SettingsLayer>['icon'];

// LayerSwitch toggles which config layer the form edits: the project's
// .gavel.yaml or the user's ~/.gavel.yaml. The Project option is offered only
// when a project is in context (global settings have no project layer).
interface Props {
  layer: SettingsLayer;
  onChange: (layer: SettingsLayer) => void;
  hasProject: boolean;
  path: string;
}

export function LayerSwitch({ layer, onChange, hasProject, path }: Props) {
  const options: SegmentedOption<SettingsLayer>[] = [
    ...(hasProject
      ? [{ id: 'project' as const, label: 'Project', icon: UiFolderGit as unknown as LayerIcon }]
      : []),
    { id: 'user' as const, label: 'User', icon: UiCog as unknown as LayerIcon },
  ];
  return (
    <div className="flex items-center gap-2.5">
      <SegmentedControl<SettingsLayer>
        value={layer}
        options={options}
        onChange={onChange}
        aria-label="Config layer"
      />
      {path && <span className="font-mono text-[11px] text-muted-foreground">{path}</span>}
    </div>
  );
}
