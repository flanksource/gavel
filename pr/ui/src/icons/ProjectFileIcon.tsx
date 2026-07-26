import { Icon, type IconProps } from '@iconify/react/offline';
import defaultFileIcon from '@iconify-icons/vscode-icons/default-file';
import defaultFolderIcon from '@iconify-icons/vscode-icons/default-folder';
import binaryIcon from '@iconify-icons/vscode-icons/file-type-binary';
import cIcon from '@iconify-icons/vscode-icons/file-type-c';
import cppIcon from '@iconify-icons/vscode-icons/file-type-cpp';
import csharpIcon from '@iconify-icons/vscode-icons/file-type-csharp';
import cssIcon from '@iconify-icons/vscode-icons/file-type-css';
import diffIcon from '@iconify-icons/vscode-icons/file-type-diff';
import dockerIcon from '@iconify-icons/vscode-icons/file-type-docker';
import gitIcon from '@iconify-icons/vscode-icons/file-type-git';
import goIcon from '@iconify-icons/vscode-icons/file-type-go';
import goWorkIcon from '@iconify-icons/vscode-icons/file-type-go-work';
import graphqlIcon from '@iconify-icons/vscode-icons/file-type-graphql';
import helmIcon from '@iconify-icons/vscode-icons/file-type-helm';
import htmlIcon from '@iconify-icons/vscode-icons/file-type-html';
import imageIcon from '@iconify-icons/vscode-icons/file-type-image';
import javaIcon from '@iconify-icons/vscode-icons/file-type-java';
import jsIcon from '@iconify-icons/vscode-icons/file-type-js';
import jsonIcon from '@iconify-icons/vscode-icons/file-type-json';
import kotlinIcon from '@iconify-icons/vscode-icons/file-type-kotlin';
import lessIcon from '@iconify-icons/vscode-icons/file-type-less';
import logIcon from '@iconify-icons/vscode-icons/file-type-log';
import makefileIcon from '@iconify-icons/vscode-icons/file-type-makefile';
import markdownIcon from '@iconify-icons/vscode-icons/file-type-markdown';
import mdxIcon from '@iconify-icons/vscode-icons/file-type-mdx';
import npmIcon from '@iconify-icons/vscode-icons/file-type-npm';
import pdfIcon from '@iconify-icons/vscode-icons/file-type-pdf2';
import phpIcon from '@iconify-icons/vscode-icons/file-type-php';
import pnpmIcon from '@iconify-icons/vscode-icons/file-type-pnpm';
import protobufIcon from '@iconify-icons/vscode-icons/file-type-protobuf';
import pythonIcon from '@iconify-icons/vscode-icons/file-type-python';
import reactJsIcon from '@iconify-icons/vscode-icons/file-type-reactjs';
import reactTsIcon from '@iconify-icons/vscode-icons/file-type-reactts';
import rubyIcon from '@iconify-icons/vscode-icons/file-type-ruby';
import rustIcon from '@iconify-icons/vscode-icons/file-type-rust';
import sassIcon from '@iconify-icons/vscode-icons/file-type-sass';
import scalaIcon from '@iconify-icons/vscode-icons/file-type-scala';
import scssIcon from '@iconify-icons/vscode-icons/file-type-scss';
import shellIcon from '@iconify-icons/vscode-icons/file-type-shell';
import sqlIcon from '@iconify-icons/vscode-icons/file-type-sql';
import svgIcon from '@iconify-icons/vscode-icons/file-type-svg';
import svelteIcon from '@iconify-icons/vscode-icons/file-type-svelte';
import swiftIcon from '@iconify-icons/vscode-icons/file-type-swift';
import terraformIcon from '@iconify-icons/vscode-icons/file-type-terraform';
import textIcon from '@iconify-icons/vscode-icons/file-type-text';
import tomlIcon from '@iconify-icons/vscode-icons/file-type-toml';
import typescriptIcon from '@iconify-icons/vscode-icons/file-type-typescript';
import vueIcon from '@iconify-icons/vscode-icons/file-type-vue';
import xmlIcon from '@iconify-icons/vscode-icons/file-type-xml';
import yamlIcon from '@iconify-icons/vscode-icons/file-type-yaml';
import zipIcon from '@iconify-icons/vscode-icons/file-type-zip';

const icons = {
  binary: binaryIcon, c: cIcon, cpp: cppIcon, csharp: csharpIcon, css: cssIcon, diff: diffIcon,
  docker: dockerIcon, file: defaultFileIcon, folder: defaultFolderIcon, git: gitIcon, go: goIcon,
  'go-work': goWorkIcon, graphql: graphqlIcon, helm: helmIcon, html: htmlIcon, image: imageIcon,
  java: javaIcon, javascript: jsIcon, json: jsonIcon, kotlin: kotlinIcon, less: lessIcon, log: logIcon,
  makefile: makefileIcon, markdown: markdownIcon, mdx: mdxIcon, npm: npmIcon, pdf: pdfIcon, php: phpIcon,
  pnpm: pnpmIcon, protobuf: protobufIcon, python: pythonIcon, 'react-js': reactJsIcon,
  'react-ts': reactTsIcon, ruby: rubyIcon, rust: rustIcon, sass: sassIcon, scala: scalaIcon, scss: scssIcon,
  shell: shellIcon, sql: sqlIcon, svg: svgIcon, svelte: svelteIcon, swift: swiftIcon,
  terraform: terraformIcon, text: textIcon, toml: tomlIcon, typescript: typescriptIcon, vue: vueIcon,
  xml: xmlIcon, yaml: yamlIcon, zip: zipIcon,
} satisfies Record<string, IconProps['icon']>;

type ProjectFileIconType = keyof typeof icons;

const extensionIcons: Record<string, ProjectFileIconType> = {
  '7z': 'zip', bash: 'shell', bat: 'shell', bin: 'binary', bz2: 'zip', c: 'c', cc: 'cpp',
  cmd: 'shell', cpp: 'cpp', cs: 'csharp', css: 'css', diff: 'diff', dll: 'binary', dylib: 'binary',
  fish: 'shell', gif: 'image', go: 'go', gql: 'graphql', graphql: 'graphql', gz: 'zip', h: 'c',
  helm: 'helm', hpp: 'cpp', htm: 'html', html: 'html', ico: 'image', java: 'java', jpeg: 'image',
  jpg: 'image', js: 'javascript', json: 'json', jsonl: 'json', jsx: 'react-js', kt: 'kotlin',
  kts: 'kotlin', less: 'less', log: 'log', md: 'markdown', mdx: 'mdx', mjs: 'javascript',
  pdf: 'pdf', php: 'php', png: 'image', proto: 'protobuf', ps1: 'shell', py: 'python', rar: 'zip',
  rb: 'ruby', rs: 'rust', sass: 'sass', scala: 'scala', scss: 'scss', sh: 'shell', so: 'binary',
  sql: 'sql', svg: 'svg', svelte: 'svelte', swift: 'swift', tar: 'zip', tf: 'terraform',
  tfvars: 'terraform', tgz: 'zip', toml: 'toml', ts: 'typescript', tsx: 'react-ts', txt: 'text',
  vue: 'vue', wasm: 'binary', webp: 'image', xml: 'xml', xz: 'zip', yaml: 'yaml', yml: 'yaml',
  zip: 'zip', zsh: 'shell',
};

const fileNameIcons: Record<string, ProjectFileIconType> = {
  '.gitattributes': 'git', '.gitignore': 'git', dockerfile: 'docker', 'go.mod': 'go', 'go.sum': 'go',
  'go.work': 'go-work', 'go.work.sum': 'go-work', makefile: 'makefile', 'package-lock.json': 'npm',
  'package.json': 'npm', 'pnpm-lock.yaml': 'pnpm', 'pnpm-workspace.yaml': 'pnpm',
};

export function ProjectFileIcon({ path, directory, className }: { path: string; directory: boolean; className?: string }) {
  const type = directory ? 'folder' : fileIconType(path);
  return <Icon icon={icons[type]} className={className} data-file-icon={type} aria-hidden="true" />;
}

function fileIconType(path: string): ProjectFileIconType {
  const fileName = path.toLowerCase().split('/').at(-1) ?? '';
  if (fileNameIcons[fileName]) return fileNameIcons[fileName];
  return extensionIcons[fileName.split('.').at(-1) ?? ''] ?? 'file';
}
