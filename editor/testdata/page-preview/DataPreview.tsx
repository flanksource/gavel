import Title from './title.jsx';
import './page.css';

type Release = { name?: string; version?: string; owners?: string[] };

export default function DataPreview({ data }: { data: Release }) {
  return (
    <main>
      <Title>{data.name ?? 'Unnamed release'}</Title>
      <p>Version {data.version ?? 'unset'}</p>
      <ul>{(data.owners ?? []).map((owner) => <li key={owner}>{owner}</li>)}</ul>
    </main>
  );
}
